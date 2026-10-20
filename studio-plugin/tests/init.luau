--> L4 plugin mapping tests (Lune). No Roblox Studio.

local process = require("@lune/process")

local mappingTests = require("@self/mapping_test")

local failed = 0
local passed = 0

for _, case in mappingTests do
	local ok, err = pcall(case.fn)
	if ok then
		passed += 1
		print("ok   " .. case.name)
	else
		failed += 1
		print("FAIL " .. case.name)
		print("     " .. tostring(err))
	end
end

print(string.format("%d passed, %d failed", passed, failed))
if failed > 0 then
	process.exit(1)
end
