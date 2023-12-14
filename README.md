# BaumanS2SBot

Telegram Bot, where students can help other students by posting requests for help

<img src="https://github.com/MMacardi/BaumanS2SBot/assets/61706774/99b67728-f346-413b-9b32-2dc6ecc6fb49.png" width="300" height="300" />

To Start Docker Container use 

`docker-compose -f docker-compose.dev.yml start`
to check docker how it is working 
`docker-compose -f docker-compose.dev.yml ps`

For easier work with database this project user goose and sqlx library

To make migration to db use 
`goose postgres "DATASOURCE_NAME" up`

