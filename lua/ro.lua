do
  local setField = function(m) m.seconds = m.seconds + 1 end

  local msg = proto.new("google.protobuf.Duration")
  print(msg:IsReadOnly())
  --> =false
  setField(msg)
  print(msg.seconds)
  --> =1

  msg = msg:ReadOnly()
  print(msg:IsReadOnly())
  --> =true
  setField(msg)
  print(msg.seconds)
  --> =1

  msg = msg:ReadOnly()
  print(msg:IsReadOnly())
  --> =true
  setField(msg)
  print(msg.seconds)
  --> =1
end

do
  local msg = proto.new("google.protobuf.ListValue")
  local values = msg.values

  print(values:IsReadOnly())
  --> =true
end

do
  local msg = proto.new("google.protobuf.Struct")
  local fields = msg.fields

  print(fields:IsReadOnly())
  --> =true
end