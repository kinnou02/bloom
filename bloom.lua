-- bloom.lua
counter = 0
function request()
	counter = counter + 1
	local key = "key-" .. (counter % 1000000) -- ou plus petit si tu veux
	return wrk.format("GET", "/check?key=" .. key)
end
