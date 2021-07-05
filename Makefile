auction:
	go run -race cmd/auction/main.go

bidder:
	go run cmd/bidder/main.go

simulator:
	go run cmd/simulator/main.go

ethnode:
	ethnode --workdir config/dev/

test:
	go test -race -v ./...
