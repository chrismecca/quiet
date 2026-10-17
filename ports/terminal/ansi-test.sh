#!/bin/sh
# @(#)quiet / sh / ports/terminal/ansi-test.sh / 2026-10-17

set -eu

name()
{
	case "$1" in
		0) printf 'black' ;;
		1) printf 'red' ;;
		2) printf 'green' ;;
		3) printf 'yellow' ;;
		4) printf 'blue' ;;
		5) printf 'magenta' ;;
		6) printf 'cyan' ;;
		7) printf 'white' ;;
	esac
}

printf '\n'
printf 'Quiet ANSI terminal reference test\n'
printf '==================================\n'

printf '\nnormal foregrounds\n\n'

i=0
while [ "$i" -lt 8 ]; do
	printf '\033[3%dm %2d %-7s \033[0m  ' \
	    "$i" \
	    "$i" \
	    "$(name "$i")"

	if [ $(((i + 1) % 4)) -eq 0 ]; then
		printf '\n'
	fi

	i=$((i + 1))
done

printf '\nbright foregrounds\n\n'

i=0
while [ "$i" -lt 8 ]; do
	printf '\033[9%dm %2d bright %-7s \033[0m  ' \
	    "$i" \
	    "$((i + 8))" \
	    "$(name "$i")"

	if [ $(((i + 1) % 4)) -eq 0 ]; then
		printf '\n'
	fi

	i=$((i + 1))
done

printf '\nnormal backgrounds\n\n'

i=0
while [ "$i" -lt 8 ]; do
	printf '\033[4%dm\033[97m %2d background \033[0m  ' \
	    "$i" \
	    "$i"

	if [ $(((i + 1) % 4)) -eq 0 ]; then
		printf '\n'
	fi

	i=$((i + 1))
done

printf '\nbright backgrounds\n\n'

i=0
while [ "$i" -lt 8 ]; do
	printf '\033[10%dm\033[30m %2d background \033[0m  ' \
	    "$i" \
	    "$((i + 8))"

	if [ $(((i + 1) % 4)) -eq 0 ]; then
		printf '\n'
	fi

	i=$((i + 1))
done

printf '\nattributes\n\n'

printf '\033[1m bold text \033[0m\n'
printf '\033[2m dim text \033[0m\n'
printf '\033[4m underlined text \033[0m\n'
printf '\033[7m reverse video \033[0m\n'

printf '\nsemantic approximations\n\n'

printf '\033[34m info / mineral \033[0m\n'
printf '\033[32m nominal / moss \033[0m\n'
printf '\033[33m warning / ochre \033[0m\n'
printf '\033[31m error / rust \033[0m\n'

printf '\n'
printf 'selection: drag across this sentence with the mouse.\n'
printf 'cursor: inspect the shell cursor after this program exits.\n'
printf '\n'

