include mk/env.mk

# make build-kj CMDTAG=xxx 构建指定服务
# make build-kj 构建所有服务
build-kj:
ifdef CMDTAG
	@go build -o ./bin/game-kj-$(CMDTAG) ./cmd/$(CMDTAG)/
else
	@for dir in $(shell ls cmd/); do \
		go build -o ./bin/game-kj-$$dir ./cmd/$$dir/; \
	done
endif

# make run-kj CMDTAG=xxx
run-kj:
	@./bin/game-kj-$(filter-out $@,$(CMDTAG))

# 编译 + 运行 build-run-kj CMDTAG=xxx
build-run-kj:
	@go build -o ./bin/game-kj-$(CMDTAG) ./cmd/$(CMDTAG)/
	@./bin/game-kj-$(CMDTAG)
