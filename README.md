# BaumanS2SBot
To Start Docker Container use 
docker-compose -f docker-compose.dev.yml start
docker-compose -f docker-compose.dev.yml ps
For easier work with database this project user goose and sqlx library
To make migration to db use goose postgres "DATASOURCE_NAME" up
