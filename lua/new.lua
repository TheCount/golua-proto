do
  local msg = proto.new("google.protobuf.Duration")

  print(msg.seconds)
  --> =0

  print(msg.nanos)
  --> =0
end

do
  local msg1 = proto.new("google.protobuf.Duration")
  local msg2 = proto.new("google.protobuf.Timestamp")
  local msg3 = proto.new("google.protobuf.Duration")

  print(msg1 == msg2)
  --> =false
  print(msg1 == msg3)
  --> =true
  print(msg1:Type() == msg2:Type())
  --> =false
  print(msg1:Type() == msg3:Type())
  --> =true
end