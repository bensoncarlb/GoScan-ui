# GoScan-UI
This is a UI package for interacting with a [GoScan](https://github.com/bensoncarlb/GoScan) server. GoScan is a barebones proof-of-concept implementation of an document imaging server with the primary functionality of data extraction. 

## Prerequisites
* go 1.26.1
* A running [GoScan](https://github.com/bensoncarlb/GoScan) process on the same system.

## Usage
```go run github.com/bensoncarlb/GoScan-ui@latest```
> The first time the project runs, fyne does a lot of initial setup in the background. This is cached for subsequent runs, but does mean the first start can take quite a while.

* Status
    * A basic check that the server's /ping can be reached.
* Documents
    * View a list of the currently processed items. Clicking on an item will open up the image and recorded data. 
* Types <br /><br />
A ```Document Type``` contains the definition for a unique form or document and how to record it. A Document Type definition includes one or more Regions. A Region is a rectangular section of pixels on a Document denoting a field. Regions are recorded as two points representing opposing corners.

    * View a list of the currently configured Document Types. Clicking on a type will delete it from the server.
    * To add a new type:
        1. Drag a png file onto the window. 
        2. Enter a Title and Identifier (lowercased). 
        3. Once a region is defined a red box outlining the region will be added. For each defined region, a box below Submit will be displayed to enter a field name for identfying the region.
        4. Once all fields are populated, Submit will save off the new Type. If successful the image will close.

## Limitations
UI design is a black magic of which I know not the art.

## Motivation
I needed a way to manage Document Types for testing development on GoScan. This is slightly less painful versus manually writing JSON elements of randomly guessed pixel regions.

## Planned Future Work

* Make the UI reasonably usable.
* More flexibility when defining regions on a new Document Type, such as deleting.
* Flags or a configuration file for controlling the limited options currently hardcoded in.