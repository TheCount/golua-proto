do
  local msg = proto.new("google.protobuf.Duration")

  print(msg:Has("seconds"))
  --> =false
  print(msg:Has(1))
  --> =false
end
