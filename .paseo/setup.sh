#!/bin/sh
set -e
# .paseo/setup.sh — Paseo chạy file này mỗi khi tạo worktree mới cho agent.
# Nguyên tắc: idempotent — chạy lại bao nhiêu lần cũng an toàn, không lỗi.
# Bản universal: tự nhận diện loại repo, không cần sửa gì khi copy sang repo khác.

# Tạo .env từ file mẫu nếu repo có .env.example mà chưa có .env
[ -f .env ] || [ ! -f .env.example ] || cp .env.example .env

# Tự nhận diện stack qua file đặc trưng rồi cài dependencies
if [ -f pnpm-lock.yaml ]; then
  pnpm install --frozen-lockfile
elif [ -f yarn.lock ]; then
  yarn install --frozen-lockfile
elif [ -f package-lock.json ]; then
  npm ci
elif [ -f package.json ]; then
  npm install
elif [ -f requirements.txt ]; then
  pip install -r requirements.txt
elif [ -f pyproject.toml ]; then
  if command -v uv >/dev/null 2>&1; then
    uv sync
  else
    pip install -e .
  fi
elif [ -f go.mod ]; then
  go mod download
elif [ -f Cargo.toml ]; then
  cargo fetch
elif [ -f Gemfile ]; then
  bundle install
elif [ -f composer.json ]; then
  composer install
else
  echo "Khong nhan dien duoc stack, bo qua cai dependencies."
fi
