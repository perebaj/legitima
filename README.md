
![Logo](/assets/blackbackground.png#gh-dark-mode-only)
![Logo](/assets/whitebackgorund.png#gh-light-mode-only)

![image](assets/legitima.png)

# Legitima

Simple API to authenticate Services

## Google Auth

The required environment variables to use the Google Auth are:

```
export LEGITIMA_GOOGLE_CLIENT_ID=
export LEGITIMA_GOOGLE_CLIENT_SECRET=
export LEGITIMA_MYSQL_URL="root:mysql@tcp(localhost:3307)/mysql" <- Example for local tests (for a while)
```


## Command Line

All commands could be accessed using: `Make help`

To reproduce the tests and lint, just run respectively: `make test` and `make lint`.

## Tests

For a while the **integration tests** are just able to run locally, so we need to start the development environment, using the command: `make dev/start`, then we can run the integration tests using the command: `make integration-test`

## Ship a new version

- `make image/publish`
- `heroku container:release web -a legitima`

## Heroku Logs

- `heroku logs --tail -a legitima`
