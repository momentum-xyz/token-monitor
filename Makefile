gen:
	abigen --sol=./solidity_contracts/erc20.sol --pkg abigen --type=ERC20 > ./pkg/abigen/erc20.go
	abigen --sol=./solidity_contracts/erc721.sol --pkg abigen --type=ERC721 > ./pkg/abigen/erc721.go
	abigen --sol=./solidity_contracts/erc1155.sol --pkg abigen --type=ERC1155 > ./pkg/abigen/erc1155.go
	abigen --sol=./solidity_contracts/multicall2.sol --pkg abigen --type=Multicall2 > ./pkg/abigen/multicall2.go --alias Call=Multicall2Call

up:
	docker compose up -d

run:
	go run ./cmd/token_service

build:
	docker build .
