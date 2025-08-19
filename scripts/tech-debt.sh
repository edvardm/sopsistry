#!/bin/bash
set -euo pipefail

echo "📌 FIXME Comments (High Priority):"
rg --type go --line-number --no-heading "FIXME|FIX ME" . || echo "  ✅ None found"
echo ""
echo "⚠️  Linter Ignores (Review Needed):"
rg --type go --line-number --no-heading "//\s*nolint" . || echo "  ✅ None found"
echo ""
echo "📝 TODO Comments (Low Priority):"
rg --type go --line-number --no-heading "TODO|TO DO" . || echo "  ✅ None found"