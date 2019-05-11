package martaapi

const ValidScheduleJSON = `
[
  {
    "DESTINATION": "Doraville",
    "DIRECTION": "N",
    "EVENT_TIME": "5/11/2019 5:48:05 PM",
    "LINE": "GOLD",
    "NEXT_ARR": "05:48:14 PM",
    "STATION": "LAKEWOOD STATION",
    "TRAIN_ID": "304326",
    "WAITING_SECONDS": "-16",
    "WAITING_TIME": "Boarding"
  },
  {
    "DESTINATION": "Hamilton E Holmes",
    "DIRECTION": "W",
    "EVENT_TIME": "5/11/2019 5:48:17 PM",
    "LINE": "BLUE",
    "NEXT_ARR": "05:48:26 PM",
    "STATION": "KENSINGTON STATION",
    "TRAIN_ID": "103206",
    "WAITING_SECONDS": "-4",
    "WAITING_TIME": "Boarding"
  }
]`

var ValidScheduleExpectation = []Schedule{
	Schedule{
		Destination:    "Doraville",
		Direction:      "N",
		EventTime:      "5/11/2019 5:48:05 PM",
		Line:           "GOLD",
		NextArrival:    "05:48:14 PM",
		Station:        "LAKEWOOD STATION",
		TrainID:        "304326",
		WaitingSeconds: "-16",
		WaitingTime:    "Boarding",
	},
	Schedule{
		Destination:    "Hamilton E Holmes",
		Direction:      "W",
		EventTime:      "5/11/2019 5:48:17 PM",
		Line:           "BLUE",
		NextArrival:    "05:48:26 PM",
		Station:        "KENSINGTON STATION",
		TrainID:        "103206",
		WaitingSeconds: "-4",
		WaitingTime:    "Boarding",
	},
}
