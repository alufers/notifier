
### /notify

#### POST
##### Summary

Send a notification

##### Description

Delivers a notification to all the sinks

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| notification | body | Notification to deliver | Yes | [notifier.PostNotifyBody](#notifierpostnotifybody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [notifier.PostNotifyResponse](#notifierpostnotifyresponse) |
| 400 | Bad Request | [notifier.ErrorResponse](#notifiererrorresponse) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### /question

#### POST
##### Summary

Asks a question to the user

##### Description

Currently supported question types: yesno

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| notification | body | Question to ask | Yes | [notifier.PostQuestionBody](#notifierpostquestionbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [notifier.PostQuestionResponse](#notifierpostquestionresponse) |
| 400 | Bad Request | [notifier.ErrorResponse](#notifiererrorresponse) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| ApiKeyAuth | |

### Models

#### notifier.Answer

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| answerDuration | integer |  | No |
| timedOut | boolean |  | No |
| value |  |  | No |

#### notifier.ErrorResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| error | string |  | No |

#### notifier.PostNotifyBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| body | string |  | No |
| title | string |  | No |

#### notifier.PostNotifyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| deliveriesCucceeded | integer |  | No |
| deliveriesTotal | integer |  | No |
| errors | object |  | No |

#### notifier.PostQuestionBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| kind | string |  | No |
| text | string |  | No |
| timeout | string |  | No |

#### notifier.PostQuestionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| answer | [notifier.Answer](#notifieranswer) |  | No |
| errors | object |  | No |
