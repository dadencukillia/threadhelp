ifneq ($(wildcard ./.env),)
    include ./.env
    export
endif

.DEFAULT_GOAL := run

run:
ifeq (true, ${USE_DBPANEL})
	docker-compose -f docker-compose.yml -f docker-compose-dbpanel.yml up --build
else
	docker-compose -f docker-compose.yml up --build
endif

stop:
	docker-compose -f docker-compose.yml -f docker-compose-dbpanel.yml down
