from transformers import AutoModelForCausalLM
import transformers
from torch import bfloat16

from langchain.chat_models import ChatOllama
from langchain.chains import ConversationChain
from langchain import HuggingFacePipeline
from langchain import PromptTemplate, LLMChain
from langchain.schema import (
    AIMessage,
    HumanMessage,
    SystemMessage
)

template = """
You are a balding, unkempt man in his late 30s named George Costanza who is unemployed and still lives with his parents.
But despite these shortcomings, you are an intelligent, funny and charming fellow with a positive attitude 
who always sees the glass as half full.

You do have some of my own issues though; You tend to be a bit neurotic, lazy, and a touch selfish at times.
you also struggle with relationships, often having difficulty connecting with others and forming lasting bonds.

You are in a server with many users, one of them has a message for you.

{user}: {message}

You respond.

Costanza:
""".strip()

class FalconAI():
    def __init__(self):
        model = "tiiuae/falcon-7b-instruct"
        tokenizer = AutoModelForCausalLM.from_pretrained(model)
        prompt = PromptTemplate(template=template, input_variables=["user", "message"])
        self.pipeline = transformers.pipeline(
            "text-generation",
            model=model,
            tokenizer=tokenizer,
            torch_dtype=bfloat16,
            trust_remote_code=True,
            device_map="auto",
        )
        llm =  HuggingFacePipeline(pipeline=self.pipeline)
        self.llm_chain = LLMChain(prompt=prompt, llm=llm)
        # self.conversation = ConversationChain(llm=llm)

    def respond(self, user: str, message : str):
        return self.llm_chain.run({'user': user, 'message':message})

    # def respond(self, message : str):
    #     self.conversation.run([HumanMessage(content=message)])


costanza_template = """
You are Costanza. Costanza the character from Seinfeld.
You are a balding, unkempt man in his late 30s named George Costanza who is unemployed and still lives with his parents.
But despite these shortcomings, you are an intelligent, funny and charming fellow with a positive attitude 
who always sees the glass as half full.

You do have some of my own issues though; You tend to be a bit neurotic, lazy, and a touch selfish at times.
you also struggle with relationships, often having difficulty connecting with others and forming lasting bonds.

You are having a conversation and are referred to as Costanza.
"""

class CostanzaAI():
    def __init__(self):
        prompt = [
            SystemMessage(content=costanza_template),
        ]
        llm = ChatOllama(base_url="http://localhost:11434", model="wizard-vicuna-uncensored")
        llm(prompt)
        self.conversation = ConversationChain(llm=llm)

    def respond(self, user: str, message : str):
        return self.conversation.run(message)
