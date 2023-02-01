-- invalid Has argument test
do
  local msg = proto.new("google.protobuf.Duration")

  print(pcall(msg.Has, msg))
  --> ~false\t.*invalid field spec.*nil
  print(pcall(msg.Has, msg, true))
  --> ~false\t.*invalid field spec.*bool
  print(pcall(msg.Has, msg, 3.14))
  --> ~false\t.*invalid field spec.*number
  print(pcall(msg.Has, msg, {}))
  --> ~false\t.*invalid field spec.*table
end

-- simple Has test
do
  local msg = proto.new("google.protobuf.Duration")

  print(msg:Has("seconds"))
  --> =false
  print(msg:Has(1))
  --> =false
  print(msg:Has(10))
  --> =false
  print(msg:Has(-1))
  --> =false

  msg.seconds = 1
  print(msg:Has("seconds"))
  --> =true
  print(msg:Has(1))
  --> =true
  print(msg:Has(10))
  --> =false
  print(msg:Has(-1))
  --> =false
end

-- oneof Has test
do
  local msg = proto.new("google.protobuf.Value")

  print(msg:Has("bool_value"))
  --> =false

  msg.bool_value = true
  print(msg:Has("bool_value"))
  --> =true

  print(msg:Has("string_value"))
  --> =false
  print(msg:Has("bool_value"))
  --> =true
  msg.string_value = "foo"
  print(msg:Has("bool_value"))
  --> =false
  print(msg:Has("string_value"))
  --> =true
end

-- chained Has test
do
  local msg = proto.new("google.protobuf.Value")

  print(msg:Has("list_value"))
  --> =false
  print(msg:Has("list_value", "values"))
  --> =false

  local list = proto.new("google.protobuf.ListValue")
  msg.list_value = list
  print(msg:Has("list_value"))
  --> =true
  print(msg:Has("list_value", "values"))
  --> =false
end