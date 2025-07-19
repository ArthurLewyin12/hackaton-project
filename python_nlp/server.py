# python_nlp/server.py

import grpc
from concurrent import futures
import time

# Importer les classes générées par gRPC
from proto import symptom_analysis_pb2
from proto import symptom_analysis_pb2_grpc

# Créer une classe pour implémenter le service
class SymptomAnalysisServiceImpl(symptom_analysis_pb2_grpc.SymptomAnalysisServiceServicer):
    """Implémente la logique du service d'analyse de symptômes."""

    def Analyze(self, request, context):
        """Reçoit une requête, analyse le texte et renvoie une réponse structurée."""
        print(f"Requête reçue avec le texte : '{request.text}'")

        # --- Logique NLP (simulation pour l'instant) ---
        # Dans une vraie application, on utiliserait ici une bibliothèque comme spaCy ou Hugging Face
        # pour extraire les entités (symptômes, durée, etc.).

        # Simulation simple :
        response = symptom_analysis_pb2.SymptomAnalysisResponse()
        if "fièvre" in request.text.lower():
            response.symptoms.add(name="Fièvre", duration="non spécifiée", intensity="non spécifiée")
        if "tête" in request.text.lower():
            response.symptoms.add(name="Mal de tête", duration="non spécifiée", intensity="aigu")
        if not response.symptoms:
             response.symptoms.add(name="Aucun symptôme reconnu", duration="", intensity="")
        # --- Fin de la logique de simulation ---

        print(f"Réponse envoyée : {response.symptoms}")
        return response

def serve():
    """Démarre le serveur gRPC."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    symptom_analysis_pb2_grpc.add_SymptomAnalysisServiceServicer_to_server(
        SymptomAnalysisServiceImpl(), server
    )
    server.add_insecure_port('[::]:50051')
    print("Serveur gRPC démarré sur le port 50051...")
    server.start()
    try:
        while True:
            time.sleep(86400)  # Tourne pendant une journée
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
