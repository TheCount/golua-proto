do
  local msg = proto.new("google.protobuf.Duration")

  print(msg.seconds)
  --> =0

  print(msg.nanos)
  --> =0
end
