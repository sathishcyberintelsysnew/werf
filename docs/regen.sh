#!/bin/bash

set -e

SOURCE_PATH="$(realpath "${BASH_SOURCE[0]}")"
PROJECT_DIR="$(dirname $SOURCE_PATH)/.."

function regen() {
  saved_dir="$PWD"
  cd "$PROJECT_DIR"
  ./go-build.sh

  # regen CLI partials, pages and sidebar
  HOME='~' werf docs --log-terminal-width=100

  # regen README partials
  readme_en_partials_dir="docs/_includes/en/readme"
  readme_ru_partials_dir="docs/_includes/ru/readme"
  rm -rf "$readme_en_partials_dir" "$readme_ru_partials_dir"
  mkdir -p "$readme_en_partials_dir" "$readme_ru_partials_dir"
  werf docs --split-readme --readme "README.md" --dest "$readme_en_partials_dir"
  werf docs --split-readme --readme "README_ru.md" --dest "$readme_ru_partials_dir"

  cd "$saved_dir"
}

function create_documentation_sidebar() {
  documentation_path="$PROJECT_DIR/docs/_data/sidebars/documentation.yml"
  cli_partial_path="$PROJECT_DIR/docs/_data/sidebars/_cli.yml"
  documentation_partial_path="$PROJECT_DIR/docs/_data/sidebars/_documentation.yml"

  cat << EOF > "$documentation_path"
# This file is generated by "regen.sh" command.
# DO NOT EDIT!

# This is your sidebar TOC. The sidebar code loops through sections here and provides the appropriate formatting.

EOF
  cat "$cli_partial_path" >> "$documentation_path"
  echo -e "\n" >> "$documentation_path"
  cat "$documentation_partial_path" >> "$documentation_path"
}

regen
create_documentation_sidebar