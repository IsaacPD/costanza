import chat_pb2
import chat_pb2_grpc

from absl import flags
from absl import app

from grpc_reflection.v1alpha import reflection
from concurrent.futures import ThreadPoolExecutor
import logging
import logging.config
import threading
from typing import Iterable

from google.protobuf.json_format import MessageToJson
import grpc

import costanza_ai

FLAGS = flags.FLAGS


class ChatServer(chat_pb2_grpc.ChatService):
    def __init__(self):
        self.costanza_ai = costanza_ai.CostanzaAI()
        self._id_counter = 0
        self._lock = threading.RLock()

    def Chat(
        self,
        request_iterator: Iterable[chat_pb2.ChatMessage],
        context: grpc.ServicerContext,
    ) -> Iterable[chat_pb2.ChatMessage]:
        context.add_callback(lambda: logging.info("ending stream"))
        for request in request_iterator:
            logging.info("Received message [%s]",MessageToJson(request))
            message = chat_pb2.ChatMessage()
            response = self.costanza_ai.respond(request.user, request.content)
            logging.info("Response: %s", response)
            message.user = "Costanza"
            message.content = response
            yield message


def main(argv) -> None:
    address = "localhost:8000"
    server = grpc.server(ThreadPoolExecutor())
    chat_pb2_grpc.add_ChatServiceServicer_to_server(ChatServer(), server)
    reflection.enable_server_reflection(
        (chat_pb2.DESCRIPTOR.services_by_name["ChatService"].full_name, reflection.SERVICE_NAME), server)
    server.add_insecure_port(address)
    server.start()
    logging.info("Server serving at %s", address)
    server.wait_for_termination()


if __name__ == "__main__":
    app.run(main)
