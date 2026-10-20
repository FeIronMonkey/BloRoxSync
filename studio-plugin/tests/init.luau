--> L4 plugin mapping, session, and Studio host tests (Lune). No Roblox Studio.

local process = require("@lune/process")

local mappingTests = require("@self/mapping_test")
local sessionTests = require("@self/session_test")
local hostTests = require("@self/host_test")
local studioTests = require("@self/studio_test")

local failed = 0
local passed = 0

local all = {}
for _, case in mappingTests do
	table.insert(all, case)
end
for _, case in sessionTests do
	table.insert(all, case)
end
for _, case in hostTests do
	table.insert(all, case)
end
for _, case in studioTests do
	table.insert(all, case)
end

for _, case in all do
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
