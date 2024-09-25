ifneq ($(wildcard ./.env),)
    include ./.env
    export
endif

.DEFAULT_GOAL := run

ifeq (true, ${USE_DBPANEL})
ADDITIONAL_YAMLS := ${ADDITIONAL_YAMLS} -f docker-compose-dbpanel.yml
endif  
ifeq (true, ${USE_HTTPS}) 
ADDITIONAL_YAMLS := ${ADDITIONAL_YAMLS} -f docker-compose-letsencrypt.yml
endif

run:
	docker-compose -f docker-compose.yml${ADDITIONAL_YAMLS} up --build

stop:
	docker-compose -f docker-compose.yml -f docker-compose-dbpanel.yml -f docker-compose-letsencrypt.yml down
