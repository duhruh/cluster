


zero:
	docker service scale clusterfights_worker=0
scale:
	docker service scale clusterfights_worker=12
rabbit:
	docker-compose -f docker-compose.prod.yml up -d rabbit-manage
update:
	docker-compose -f docker-compose.prod.yml build worker
	docker-compose -f docker-compose.prod.yml push worker
	docker service update --image 127.0.0.1:5000/cluster/worker clusterfights_worker
file:
	docker-compose -f docker-compose.prod.yml up file-server
manager:
	docker-compose up manager

