-- test msg:FullName
do
  local msg = proto.new("google.protobuf.Duration")
  print(msg:FullName())
  --> =google.protobuf.Duration
end

-- test msg:Name
do
  local msg = proto.new("google.protobuf.Duration")
  print(msg:Name())
  --> =Duration
end