#!/bin/bash
 
# 删除远端的tag
# 参数1: 远端仓库名称，默认为origin
remote=${https://github.com/yanglunara/discovery.git:-origin}
 
# 获取所有本地tag
local_tags=$(git tag)
 
# 遍历所有本地tag并删除远端对应的tag
for tag in $local_tags; do
    if git push $remote --delete $tag; then
        echo "Tag '$tag' deleted successfully."
    else
        echo "Failed to delete tag '$tag'."
    fi
done
