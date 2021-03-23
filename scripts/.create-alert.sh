issue_count=$(curl -H "Accept: application/vnd.github+json" -H "Authorization: Bearer $(GITHUBTOKEN)" https://api.github.com/repos/Gregory-Pereira/alerts/issues | jq length -)
sed -i '' -e "s/Test number REPACE_ME/Test number $issue_count/" .alert
sed -i '' -e "s/Groupkey number REPLACE_GROUP_KEY/Groupkey number $issue_count/" .alert
msg="$(cat ./.alert)"
echo $msg
sed -i '' -e "s/Test number $issue_count/Test number REPACE_ME/" .alert
sed -i '' -e "s/Groupkey number $issue_count/Groupkey number REPLACE_GROUP_KEY/" .alert
curl -XPOST --data-binary "${msg}" http://localhost:9393/v1/receiver