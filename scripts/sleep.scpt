set isLocked to false
set isScreensaverRunning to false

tell application "System Events"
	-- Check if login window is visible (indicates locked screen)
	if exists process "loginwindow" then
		tell process "loginwindow"
			if (count of windows) > 0 then
				set isLocked to true
			end if
		end tell
	end if
end tell

-- Check if screensaver is running
set isScreensaverRunning to (do shell script "pgrep -q ScreenSaverEngine && echo 1 || echo 0") = "1"

-- Check if Mac is asleep
set isAsleep to (do shell script "pmset -g | grep 'sleeping' | wc -l | tr -d ' '") = "1"

set jsonOutput to "{\"isLocked\": " & isLocked & ", \"isScreensaverRunning\": " & isScreensaverRunning & ", \"isAsleep\": " & isAsleep & "}"

return jsonOutput
