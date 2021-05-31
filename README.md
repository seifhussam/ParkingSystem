# ParkingSystem
This is a parking system that tracks updates of a registered parking, both virtually and physically. The former is done through users of a chatbot service that -upon request- finds the nearest available spot from an input entrance number (in case a parking spot has multiple entrances), and -upon confirmation- reserves this spot for the user who demands it. 

For simplicity purposes, we only have a prototype of one parking map with only 4 parking spaces. Each parking space has an object detection sensor and RGB LED that indicates its current status [Available - Not Available - Pending/Reserved]. 

The parking sensors are connected to the Arduino, which every 5 seconds sends a request to the server with latest updates of the parking slots, and the server replies to the Arduino the requests for reserving a spot (if any). 

In case we have a request for a certain spot the RGB LED turns blue for approximately 2 minutes, to indicate that this is a reserved spot. If the driver does not show up within 2 minutes, the parking spot is, again, made available for others, the LED indicator turns green, and the arduino thus, updates the server in its following request.

* SayeS --> Arduino Code 
* SayeSnode --> Nodemcu Code 
* main.go --> Telegram chatbot developed in golang; handle @sayes_bot 

Screenshots and Videos 
- https://drive.google.com/open?id=1TyF-ATQqPHLsRX1EavscOz8ocWFy-4Hf (demo starts at 2:40) 
