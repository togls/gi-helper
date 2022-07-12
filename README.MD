# GI-helper
sign helper for mihoyo bbs and cloud gmae

## docker compose

```yaml
version: "3"
services:
  gi:
    image: gcr.io/togls/gi-helper:0.0.1
    volumes:
      - ./your/config.json:/etc/gihelper/config.json
    environment:
      - TZ=Asia/Shanghai # default timezone
      - LOG_LEVEL=info # default log level
      - CHECK_TIME=06:00 # default check time
```

## Credits

[genshin-impact-helper](https://github.com/y1ndan)