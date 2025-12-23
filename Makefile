docker-update:
	git pull
	docker compose up -d --build
	docker compose logs -f
