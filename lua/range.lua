do
  local msg = proto.new("google.protobuf.ListValue")
  local values = msg.values

  for i, v in values:Range() do
    print(i)
  end
  -- for loop over empty should not have any output
  --> =
end