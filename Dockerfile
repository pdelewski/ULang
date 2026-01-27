# ULang/goany Development Environment
# Supports Go, C++, C#, Rust, and JavaScript backends

FROM ubuntu:22.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Detect architecture for downloading correct Go binary
ARG TARGETARCH

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    g++ \
    clang-format \
    pkg-config \
    libgl1-mesa-dev \
    libx11-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.24 (architecture-aware)
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "arm64" ]; then \
        GO_ARCH="arm64"; \
    else \
        GO_ARCH="amd64"; \
    fi && \
    wget https://go.dev/dl/go1.24.0.linux-${GO_ARCH}.tar.gz && \
    tar -C /usr/local -xzf go1.24.0.linux-${GO_ARCH}.tar.gz && \
    rm go1.24.0.linux-${GO_ARCH}.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install .NET SDK 9.0 using official install script (works on ARM64 and AMD64)
RUN wget https://dot.net/v1/dotnet-install.sh -O dotnet-install.sh \
    && chmod +x dotnet-install.sh \
    && ./dotnet-install.sh --channel 9.0 --install-dir /usr/share/dotnet \
    && rm dotnet-install.sh \
    && ln -s /usr/share/dotnet/dotnet /usr/bin/dotnet

ENV DOTNET_ROOT=/usr/share/dotnet
ENV PATH="${PATH}:/usr/share/dotnet"

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup component add rustfmt

# Install Node.js 20.x
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /ulang

# Copy project files
COPY . .

# Build astyle library (required for CGO)
RUN cd compiler/astyle && make -f Makefile.cgo clean && make -f Makefile.cgo

# Build the goany compiler
RUN cd cmd && go build -o goany .

# Add cmd to PATH for easy access
ENV PATH="/ulang/cmd:${PATH}"

# Default command - show help
CMD ["goany", "--help"]
