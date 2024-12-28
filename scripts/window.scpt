tell application "System Events"
    tell process "%s"
        if (count of windows) > 0 then
            get name of front window
        else
            ""
        end if
    end tell
end tell
