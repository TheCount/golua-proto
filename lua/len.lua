do
  local msg = proto.new("google.protobuf.ListValue")
  local values = msg.values

  print(#values)
  --> =0
end