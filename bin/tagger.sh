#!/usr/bin/env bash

TAG_PATH=$HOME/.local/acme/tags

maybe_apply_tag() {
	winid=$1
	check=$2
	tag=$3

	existing_tag=$(9p read acme/$winid/tag)

	if [[ "$existing_tag$" != *"$check"* ]]; then
		echo -n " $tag" | 9p write acme/$winid/tag
	fi
}

check_tag() {
	winid=$1
	filename=$(basename -- "$2")
	extension="${filename##*.}"

	tag_file="${TAG_PATH}/${extension}"

	if [[ -f "$tag_file" ]]; then
		check=$(cat $tag_file | cut -d ':' -f 1)
		tag=$(cat $tag_file | cut -d ':' -f 2)

		maybe_apply_tag $winid $check "$tag"
	fi
}

9p read acme/log | while read CMD; do
	winid=$(echo $CMD | cut -d ' ' -f 1)
	event=$(echo $CMD | cut -d ' ' -f 2)
	name=$(echo $CMD | cut -d ' ' -f 3)

	case $event in
		new|put)
			check_tag $winid "$name"
			;;
	esac

	# event = put, focus, new, get, del
	# Interested in new (no name) and put (named)
	# new can have a name if it is opening an existing file
done

