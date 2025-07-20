# python_nlp/server.py

import grpc
from concurrent import futures
import time
import spacy

# Importer les classes générées par gRPC
from proto import symptom_analysis_pb2
from proto import symptom_analysis_pb2_grpc

# Charger le modèle spaCy pour le français
try:
    nlp = spacy.load("fr_core_news_sm")
except OSError:
    print("Modèle spaCy 'fr_core_news_sm' non trouvé. Veuillez l'installer avec :")
    print("python -m spacy download fr_core_news_sm")
    exit()

# Mots-clés pour identifier les racines de symptômes pertinents
SYMPTOM_ROOT_KEYWORDS = {"douleur", "fièvre", "toux", "gorge", "tête", "fatigue", "courbature", "nausée", "vertige"}
NEGATION_WORDS = {"pas", "sans", "aucune", "non"}

class SymptomAnalysisServiceImpl(symptom_analysis_pb2_grpc.SymptomAnalysisServiceServicer):
    """Implémente la logique du service d'analyse de symptômes."""

    def Analyze(self, request, context):
        """Analyse le texte en utilisant l'analyse de dépendances pour une extraction précise."""
        print(f"Requête reçue avec le texte : '{request.text}'")
        
        doc = nlp(request.text.lower())
        response = symptom_analysis_pb2.SymptomAnalysisResponse()
        
        processed_tokens = set() # Pour ne pas traiter un mot deux fois

        for token in doc:
            if token in processed_tokens:
                continue

            # 1. Identifier un mot racine de symptôme
            if token.lemma_ in SYMPTOM_ROOT_KEYWORDS:
                symptom_name = token.lemma_
                original_phrase = [token.text]
                intensity = []
                is_negated = False

                # 2. Chercher les compléments (ex: "à la tête")
                for child in token.children:
                    if child.dep_ == "prep": # Préposition (à, de, dans...)
                        for sub_child in child.children:
                            if sub_child.dep_ == "pobj": # Objet de la préposition
                                symptom_name += f" {child.text} {sub_child.text}"
                                original_phrase.append(child.text)
                                original_phrase.append(sub_child.text)
                                processed_tokens.add(sub_child)
                        processed_tokens.add(child)

                    # 3. Chercher les adjectifs (intensité)
                    if child.dep_ == "amod":
                        intensity.append(child.text)
                        original_phrase.insert(0, child.text) # Insérer avant le nom
                        processed_tokens.add(child)

                # 4. Vérifier la négation (attachée au symptôme principal)
                if any(c.dep_ == "neg" and c.lemma_ in NEGATION_WORDS for c in token.children):
                    is_negated = True
                if token.i > 0 and doc[token.i - 1].lemma_ in NEGATION_WORDS:
                    is_negated = True

                # Ajouter le symptôme trouvé
                symptom = response.symptoms.add()
                symptom.name = symptom_name
                symptom.original_text = " ".join(original_phrase)
                symptom.negated = is_negated
                symptom.intensity = ", ".join(intensity) if intensity else "non spécifiée"
                symptom.duration = "non spécifiée"
                
                processed_tokens.add(token)

        if not response.symptoms:
            print("Aucun symptôme pertinent n'a été reconnu.")
        
        print(f"Réponse envoyée : {response.symptoms}")
        return response

def serve():
    """Démarre le serveur gRPC."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    symptom_analysis_pb2_grpc.add_SymptomAnalysisServiceServicer_to_server(
        SymptomAnalysisServiceImpl(), server
    )
    server.add_insecure_port('[::]:50051')
    print("Serveur gRPC (NLP) démarré sur le port 50051...")
    server.start()
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()