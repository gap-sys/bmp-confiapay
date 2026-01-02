all: build

build:
	docker build -t bmpconfiapay .
	docker run --name bmp-fgts -p 8080:8080 bmp

upload:
	docker tag bmpconfiapay sergiomsa/bmpconfiapay:v1
	docker push sergiomsa/bmpconfiapay:v1

upload_homolog:
	docker tag hbmpconfiapay sergiomsa/hbmpconfiapay:v1
	docker push sergiomsa/hbmpconfiapay:v1

local:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "master" ]; then \
		echo "Você está na branch '$$(git rev-parse --abbrev-ref HEAD)'. Volte para 'master'!"; \
		exit 1; \
	fi
	docker build -t bmpconfiapay .
	docker run --name bmpconfiapay -p 8080:8080 --network="host" bmpconfiapay


homolog:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "homologa" ]; then \
		echo "Você está na branch '$$(git rev-parse --abbrev-ref HEAD)'. Volte para 'homologa'!"; \
		exit 1; \
	fi
	docker build -t hbmpconfiapay -f Dockerfile.homolog  .
	docker run --name hbmpconfiapay -p 8080:8080 --network="host" hbmpconfiapay
	
