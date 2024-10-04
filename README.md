Примеры запросов:

  1) GET "http://localhost:8080/auth/token?user_guid=1234743323" - Возвращает выданную сервером пару Access- и Refresh- токенов
    Response: {"accessToken":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaXAiOiIxMjcuMC4wLjEiLCJleHAiOjE3MjgwNDk4OTksInVzZXJfZ3VpZC
                I6IjEyMzQ3NDMzMjMifQ.k6OUl6FTHgy0Uk_ZL926Izuq_fTlj9vIEqwzkCCN5LoP9KcgrTcw7RQKymwLoM9F4FAI5jOEhE9jk4PzmOqSkQ","refreshToken":"KRg6Hpnu
                I+Xe9qlKJ/9+6N6WO5p2NmrqzmMc8MUv6RuJjRYQv1RDfFFnF2da6dHROmMkzCaf|MTIzNDc0MzMyMw==|MTI3LjAuMC4x"}

  2) GET "http://localhost:8080/auth/refresh?refresh_token=KRg6HpnuI+Xe9qlKJ/9+6N6WO5p2NmrqzmMc8MUv6RuJjRYQv1RDfFFnF2da6dHROmMkz
          Caf|MTIzNDc0MzMyMw==|MTI3LjAuMC4x" -H "Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaXAiOiIxMjcuMC4wLjEiLCJl
          eHAiOjE3MjgwNDk4OTksInVzZXJfZ3VpZCI6IjEyMzQ3NDMzMjMifQ.k6OUl6FTHgy0Uk_ZL926Izuq_fTlj9vIEqwzkCCN5LoP9KcgrTcw7RQKymwLoM9F4FAI5jOEhE9jk4PzmOqSkQ"
    Response: {"accessToken":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaXAiOiIxMjcuMC4wLjEiLCJleHAiOjE3MjgwNTAzNzMsInVzZXJfZ3VpZC
                I6IjEyMzQ3NDMzMjMifQ.5qgmXay-ZRNU3YoztUf8N7X1EMGj3Gypfyukdo7Ajjy5qM7R5mV6b6sHdrbqvc5k8jWXu4FvG5V1oH9vmO1NeQ","refreshToken":
                "W97V0wG 9ZDaI53WGzKmTa/lcH9lIti9 rrRYE9UTc6epHt6y0CrMgGNUCHMk7eVhu6r5Y7r|MTIzNDc0MzMyMw==|MTI3LjAuMC4x","ipChanged":false}
   
*** Если Access-токен истек, пользуясь Payload'ом Refresh-токена выдаст новый Access-токен.

*** Если Refresh-токен истек, удалит его. Чтобы получить новый нужно будет выполнить запрос из п.1.

*** Докер-Контейнер с БД сам создаёт таблицу при запуске, после закрытия контейнера данные в таблице сохраняются.

*** В файле .env хранятся переменные окружения. Переменная REFRESH_TOKEN_ROTATION_A_R отвечает за настройку ротации Refresh-токена, если ее нет в .env - то она по умолчанию FALSE, т.е при refresh-запросе Refresh-токен остается прежним, если TRUE, то при каждом refresh запросе выдается новый Refresh-токен.

*** Email Waining отправляется на Mailtrap — сервис безопасного тестирования писем.

*** Так как bcrypt хэширование не поддерживает размеры "пароля" больше 72 байтов, то хэши Refresh-токенов в БД хранятся смёрджеными bcrypt частями, обьединенными в одну строку
