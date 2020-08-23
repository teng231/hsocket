run-client:
	cd client && yarn && yarn start
run-ws:
	go build && ./hsocket -port=:3001
gen:
	cd header && ./builder.sh