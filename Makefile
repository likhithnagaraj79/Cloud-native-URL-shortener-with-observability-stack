DOCKER_COMPOSE = deployments/docker/docker-compose.yml

.PHONY: build test run docker-up docker-down logs ps clean

build:
	mvn package -DskipTests -B

test:
	mvn verify -B

run:
	mvn spring-boot:run

docker-up:
	docker compose -f $(DOCKER_COMPOSE) up --build -d

docker-down:
	docker compose -f $(DOCKER_COMPOSE) down -v

logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f app

ps:
	docker compose -f $(DOCKER_COMPOSE) ps

clean:
	mvn clean
