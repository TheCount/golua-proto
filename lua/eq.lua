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
  print(msg1 == msg1:Type(), msg1:Type() == msg1)
  --> =false	false

  msg1.seconds = 1
  print(msg1 == msg3)
  --> =false
  msg3.seconds = 1
  print(msg1 == msg3)
  --> =true
end
