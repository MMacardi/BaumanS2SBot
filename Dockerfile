# Используйте официальный образ Go как базовый
FROM golang:1.16

# Установите рабочий каталог внутри контейнера
WORKDIR /BaumanS2SBot

# Скопируйте файлы go.mod и go.sum, если они есть
COPY go.mod go.sum ./

# Загрузите зависимости. Это может улучшить кэширование слоев Docker
RUN go mod download

# Скопируйте исходный код бота в контейнер
COPY . .

# Скомпилируйте приложение для продакшена
RUN go build -o BaumanS2SBot ./cmd/app/main.go

RUN chmod +x BaumanS2SBot

# Запустите бота
CMD [ "./BaumanS2SBot" ]