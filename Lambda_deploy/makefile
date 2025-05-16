# ----- Makefile -------------------------------------------------
# 1) *.go を列挙
SRCS      := $(wildcard *.go)

# 2) 出力ディレクトリ（拡張子を除いた名前）
OUT_DIRS  := $(patsubst %.go,dist/%,$(SRCS))

# ----------------[ 既定ターゲット ]----------------
.PHONY: all
all: $(OUT_DIRS:=/bootstrap)

# ----------------[ ビルドルール ]-------------------
# dist/<name>/bootstrap  の依存は  <name>.go
dist/%/bootstrap: %.go
	@mkdir -p $(dir $@)          # 出力先ディレクトリを用意
	GOOS=linux GOARCH=arm64 \
		go build -tags lambda.norpc -o $@ $<

# ----------------[ お掃除 ]--------------------------
.PHONY: clean
clean:
	rm -rf dist
# ----------------------------------------------------
