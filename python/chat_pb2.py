# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chat.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\nchat.proto\x12\x08\x63ostanza\",\n\x0b\x43hatMessage\x12\x0c\n\x04user\x18\x01 \x01(\t\x12\x0f\n\x07\x63ontent\x18\x02 \x01(\t2G\n\x0b\x43hatService\x12\x38\n\x04\x43hat\x12\x15.costanza.ChatMessage\x1a\x15.costanza.ChatMessage(\x01\x30\x01\x42(Z&github.com/isaacpd/costanza/proto/chatb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'chat_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z&github.com/isaacpd/costanza/proto/chat'
  _globals['_CHATMESSAGE']._serialized_start=24
  _globals['_CHATMESSAGE']._serialized_end=68
  _globals['_CHATSERVICE']._serialized_start=70
  _globals['_CHATSERVICE']._serialized_end=141
# @@protoc_insertion_point(module_scope)
