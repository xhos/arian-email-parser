{pkgs, ...}: {
  packages = with pkgs; [
    buf
    air
    acme-sh
  ];

  languages.go.enable = true;

  scripts.run.exec = ''
    go run cmd/main.go
  '';

  scripts.build.exec = ''
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.1")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    
    echo "Building arian-parser $VERSION..."
    go build -ldflags "\
      -X arian-parser/internal/version.BuildTime=$BUILD_TIME \
      -X arian-parser/internal/version.GitCommit=$GIT_COMMIT \
      -X arian-parser/internal/version.GitBranch=$GIT_BRANCH \
    " -o arian-parser ./cmd
  '';

  scripts.version.exec = ''
    go run -ldflags "\
      -X arian-parser/internal/version.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
      -X arian-parser/internal/version.GitCommit=$(git rev-parse HEAD 2>/dev/null || echo "unknown") \
      -X arian-parser/internal/version.GitBranch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") \
    " ./cmd --version-full
  '';

  scripts.fmt.exec = ''
    go fmt ./...
  '';

  scripts.bump-proto.exec = ''
    git -C proto fetch origin
    git -C proto checkout main
    git -C proto pull --ff-only
    git add proto
    git commit -m "⬆️ bump proto files"
    git push
  '';

  git-hooks.hooks = {
    gotest.enable = true;
    gofmt.enable = true;
    govet.enable = true;
  };

  dotenv.enable = true;
}
