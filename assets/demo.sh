#!/usr/bin/env bash
# Self-animating, privacy-safe reproduction of the clean-me TUI for the README
# demo GIF. Uses illustrative/controlled data only — it never scans or touches
# the real filesystem. Colours and layouts mirror internal/tui/*.go 1:1.
#
# Rendered to assets/demo.gif via `vhs assets/demo.tape`.
set -u

# --- palette (truecolor, matches internal/tui/theme.go) ---
TITLE=$'\033[1;38;2;0;215;255m'   # accent 45  cyan, bold
DIM=$'\033[38;2;128;128;128m'      # dim 244    grey
GREEN=$'\033[1;38;2;0;215;135m'    # safe 42    green, bold (ribbon)
CHECK=$'\033[38;2;0;215;135m'      # safe 42    green
BIG=$'\033[1;38;2;255;175;0m'      # optin 214  orange, bold (reclaimable)
ORANGE=$'\033[38;2;255;175;0m'
RED=$'\033[38;2;255;95;95m'        # danger 203 red
R=$'\033[0m'

RIBBON="${GREEN}● read-only — nothing is deleted until you confirm${R}"
cls() { printf '\033[H\033[2J'; }
nl()  { printf '\n'; }
pause() { sleep "$1"; }

# ---------------------------------------------------------------- Welcome
cls
printf '  %sclean-me%s\n' "$TITLE" "$R"
printf '  %sa calm, read-only disk cleanup%s\n\n' "$DIM" "$R"
printf '  %sNothing is ever deleted until you review the full list and confirm.%s\n' "$GREEN" "$R"
printf '  %sFirst it only looks — a scan touches nothing.%s\n\n' "$DIM" "$R"
printf '  %sIT FINDS%s\n' "$DIM" "$R"
printf '    %sDocker images & volumes · package caches (npm, pip, brew)%s\n' "$DIM" "$R"
printf '    %sXcode derived data & simulators · node_modules · large files%s\n\n' "$DIM" "$R"
printf '  %sTIERS  %s%ssafe%s %s·%s %sopt-in%s %s·%s %sdanger%s   %s(danger is never preselected)%s\n' \
  "$DIM" "$R" "$CHECK" "$R" "$DIM" "$R" "$ORANGE" "$R" "$DIM" "$R" "$RED" "$R" "$DIM" "$R"
printf '    %s→ Trash%s%s is recoverable; a few caches purge %s%spermanently%s%s.%s\n\n' \
  "$CHECK" "$R" "$DIM" "$R" "$RED" "$R" "$DIM" "$R"
printf '  %s▸ Scan my system%s\n\n' "$TITLE" "$R"
printf '  %senter scan · ? help · q quit%s\n' "$DIM" "$R"
pause 2.1

# ---------------------------------------------------------------- Scanning
spin=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
for i in $(seq 0 15); do
  cls
  printf '  %sclean-me%s   %s\n\n' "$TITLE" "$R" "$RIBBON"
  printf '  %s scanning — just looking, nothing is touched…%s\n\n' "${spin[$((i % 10))]}" "$R"
  printf '  %sq quit%s\n' "$DIM" "$R"
  pause 0.1
done

# ---------------------------------------------------------------- Dashboard
cls
printf '  %sclean-me%s   %s\n\n' "$TITLE" "$R" "$RIBBON"
printf '  %s[########··············] 412 GB used · 88 GB free of 500 GB%s\n\n' "$DIM" "$R"
printf '  %s6.2 GB%s%s  reclaimable%s   %s9%s%s opportunities%s\n\n' \
  "$BIG" "$R" "$DIM" "$R" "$TITLE" "$R" "$DIM" "$R"
printf '  %-9s %s\n' "Node"   "3.6 GB"
printf '  %-9s %s\n' "Docker" "1.6 GB"
printf '  %-9s %s\n' "Xcode"  "620 MB"
printf '  %-9s %s\n' "Caches" "180 MB"
printf '  %-9s %s\n' "System" "41 MB"
printf '  %-9s %saudit only%s\n' "Git"   "$DIM" "$R"
printf '  %-9s %saudit only%s\n\n' "Redis" "$DIM" "$R"
printf '  %s▸ Smart Cleanup%s%s  pick the safe items for me%s\n' "$TITLE" "$R" "$DIM" "$R"
printf '    Explore\n    Space Explorer\n    History\n\n'
printf '  %s↑↓ move · enter select · ? help · q quit%s\n' "$DIM" "$R"
pause 2.2

# ---------------------------------------------------------------- Review
cls
printf '  %sReview — exactly what will be cleaned%s\n\n' "$TITLE" "$R"
printf '  6 items · %s6.2 GB%s · %s5 → Trash (recoverable)%s · %s1 permanent%s\n\n' \
  "$BIG" "$R" "$CHECK" "$R" "$RED" "$R"
review_item() { printf '    %8s  %s  %s\n' "$1" "$2" "$3"; }
printf '  %sNODE%s\n' "$DIM" "$R"; pause 0.12
review_item "3.6 GB" "${CHECK}→ Trash  ${R}" "~/work/api/node_modules"; pause 0.12
review_item "180 MB" "${CHECK}→ Trash  ${R}" "~/work/web/node_modules"; pause 0.12
review_item "41 MB"  "${CHECK}→ Trash  ${R}" "npm cache"; pause 0.12
printf '  %sDOCKER%s\n' "$DIM" "$R"; pause 0.12
review_item "1.6 GB" "${RED}permanent${R}" "docker build cache"; pause 0.12
printf '  %sXCODE%s\n' "$DIM" "$R"; pause 0.12
review_item "620 MB" "${CHECK}→ Trash  ${R}" "~/Library/Developer/Xcode/DerivedData"; pause 0.12
printf '  %sCACHES%s\n' "$DIM" "$R"; pause 0.12
review_item "180 MB" "${CHECK}→ Trash  ${R}" "~/Library/Caches/Homebrew"; pause 0.12
nl
printf '  %senter confirm · esc back · q quit%s\n' "$DIM" "$R"
pause 1.9

# ---------------------------------------------------------------- Cleaning / results
done_item() { printf '  %s✓ %s%s  %s%s%s\n' "$CHECK" "$R" "$1" "$DIM" "$2" "$R"; }
items_path=("~/work/api/node_modules" "~/work/web/node_modules" "npm cache" "docker build cache" "~/Library/Developer/Xcode/DerivedData" "~/Library/Caches/Homebrew")
items_note=("3.6 GB → trash" "180 MB → trash" "41 MB → trash" "1.6 GB" "620 MB → trash" "180 MB → trash")
for n in 1 2 3 4 5 6; do
  cls
  printf '  %sclean-me · cleaning system%s\n\n' "$TITLE" "$R"
  for k in $(seq 0 $((n - 1))); do done_item "${items_path[$k]}" "${items_note[$k]}"; done
  if [ "$n" -lt 6 ]; then
    printf '\n  %s cleaning (%d/6)…%s\n' "${spin[$((n % 10))]}" "$n" "$R"
    printf '  %sq quit%s\n' "$DIM" "$R"
  fi
  pause 0.22
done

printf '\n  %s%s%s\n\n' "$DIM" "──────────────────────────────────────────────────" "$R"
printf '  %sCleanup completed!%s\n\n' "$TITLE" "$R"
printf '  Freed space      %s6.2 GB%s\n' "$BIG" "$R"
printf '  Moved to Trash   5 items\n'
printf '  Failed           0\n'
printf '  Disk free        88 GB → %s94 GB%s  (%s+6.2 GB%s)\n' "$CHECK" "$R" "$CHECK" "$R"
pause 2.4
