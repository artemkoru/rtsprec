# Spec Delta

## MODIFIED Requirements

### Requirement: Configure camera sources
The system SHALL allow adding, editing, and removing RTSP camera definitions (name, URL, username, password) through web UI.

#### Scenario: Add camera
- **WHEN** admin enters camera details in web form (name, RTSP URL, credentials)
- **THEN** system saves configuration and begins capturing from that camera

#### Scenario: Add camera from the form
- **WHEN** admin fills in name and RTSP URL in the Cameras tab and presses Add
- **THEN** the page sends the camera to the server, clears the form, and shows the new camera in the list without reloading the page

#### Scenario: Required fields missing
- **WHEN** admin presses Add with an empty name or RTSP URL
- **THEN** the form marks the missing field as required and no request is sent

#### Scenario: Server rejects the camera
- **WHEN** the server returns an error for the new camera
- **THEN** the page shows the error next to the form and keeps the entered values

#### Scenario: Validate RTSP connection
- **WHEN** camera is added
- **THEN** system tests connection and reports success/failure to UI

#### Scenario: Edit camera
- **WHEN** admin modifies camera settings
- **THEN** system updates configuration and may restart capture if URL/credentials change

#### Scenario: Remove camera
- **WHEN** admin deletes camera configuration
- **THEN** system stops capturing and closes RTSP connection

#### Scenario: Multiple cameras
- **WHEN** system has 1-2 cameras configured
- **THEN** UI displays list of cameras with status (online/offline) and edit/delete options
