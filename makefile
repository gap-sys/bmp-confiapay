all: build

build:
	docker build -t bmp .
	docker run --name bmp-fgts -p 8080:8080 bmp

upload:
	docker tag bmp sergiomsa/bmp:v1
	docker push sergiomsa/bmp:v1

upload_homolog:
	docker tag hbmp sergiomsa/hbmp:v1
	docker push sergiomsa/hbmp:v1

local:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "master" ]; then \
		echo "Você está na branch '$$(git rev-parse --abbrev-ref HEAD)'. Volte para 'master'!"; \
		exit 1; \
	fi
	docker build -t bmp .
	docker run --name bmp -p 8080:8080 --network="host" bmp


homolog:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "homologa" ]; then \
		echo "Você está na branch '$$(git rev-parse --abbrev-ref HEAD)'. Volte para 'homologa'!"; \
		exit 1; \
	fi
	docker build -t hbmp -f Dockerfile.homolog  .
	docker run --name hbmp -p 8080:8080 --network="host" hbmp
	
