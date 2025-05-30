{
  languages.go.enable = true;

  scripts.run.exec = ''
    go run cmd/main.go
  '';

  scripts.fmt.exec = ''
    go fmt ./...
  '';

  dotenv.enable = true;
}
