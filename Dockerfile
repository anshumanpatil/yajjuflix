FROM jellyfin/jellyfin:latest

# Environment configuration
ENV DEBIAN_FRONTEND=noninteractive \
    LC_ALL=en_US.UTF-8 \
    LANG=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    HEALTHCHECK_URL=http://localhost:8096/health \
    JELLYFIN_DATA_DIR=/config \
    JELLYFIN_CACHE_DIR=/cache \
    JELLYFIN_CONFIG_DIR=/config/config \
    JELLYFIN_LOG_DIR=/config/log \
    JELLYFIN_WEB_DIR=/jellyfin/jellyfin-web \
    JELLYFIN_FFMPEG=/usr/lib/jellyfin-ffmpeg/ffmpeg \
    XDG_CACHE_HOME=/cache \
    MALLOC_TRIM_THRESHOLD_=131072 \
    NVIDIA_VISIBLE_DEVICES=all \
    NVIDIA_DRIVER_CAPABILITIES=compute,video,utility \
    LD_PRELOAD=/usr/lib/jellyfin/libjemalloc.so.2

# Exposed ports
EXPOSE 8096

# Container volumes
VOLUME ["/config", "/cache", "/media"]

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f "${HEALTHCHECK_URL}" || exit 1