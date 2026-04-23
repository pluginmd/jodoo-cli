# Jodoo API - Error Codes

## General Error Codes

| Code | Message | Description |
|---|---|---|
| 1 | Invalid Request | The request is invalid |
| 8017 | Invalid Company/Team | Company/Team doesn't exist or has been closed |
| 8301 | Invalid Authorization | Failed to verify the API key |
| 8302 | No Permission | No permission for API calls |
| 8303 | Company/Team Request Limit Exceeded | Call frequency exceeds limit |
| 8304 | Request Limit Exceeded | Call frequency exceeds limit |
| 9004 | Creation Failed | Failed to create the task |
| 9007 | Lock Obtain Failed | Failed to obtain the lock |
| 17017 | Invalid Parameter(s) | Invalid parameter(s) |
| 17018 | Invalid API Key | The API key is invalid |
| 17025 | Invalid transaction_id | Invalid transaction_id parameter(s) |
| 17026 | Duplicate transaction_id | The transaction_id already exists |
| 17027 | Uploading Failed | Failed to upload API file |
| 17032 | Invalid Field Type | Field type isn't supported |
| 17034 | Invalid Sub-Field Type | Sub-field type isn't supported |

## Member Error Codes

| Code | Message | Description |
|---|---|---|
| 1005 | Duplicate Email | Email already exists |
| 1010 | Member Not Found | Member does not exist |
| 1017 | Invalid Member Name | Invalid name format |
| 1018 | Invalid Email | Invalid email format |
| 1019 | Invalid Member Nickname | Invalid nickname format |
| 1022 | Invalid Company/Team | Member's company/team doesn't exist |
| 1024 | Invalid Mobile Number | Invalid mobile number |
| 1027 | Duplicate Phone Number | Phone number already exists |
| 1058 | No Permission | No permission for task list |
| 1065 | Duplicate Member | Member already in team |
| 1082 | Phone Number/Email Required | Required field missing |
| 1085 | Nickname Required | Nickname is required |
| 1087 | Duplicate Unique Fields | Duplicate unique fields |
| 1092 | Username Limit Exceeded | Username too long |
| 1096 | Invalid Request Parameter(s) | Invalid member parameter(s) |

## Role Error Codes

| Code | Message | Description |
|---|---|---|
| 1201 | Invalid Role | Role does not exist |
| 1203 | Invalid Role Group/Role | Invalid role group/role info |
| 1205 | Role Group Required | Role group is required |
| 1206 | Invalid Role Group | Role group does not exist |
| 1207 | Role Group Name Limit Exceeded | Name too long |
| 1208 | Role Group with Member(s) | Can't delete group with members |

## App & Form Error Codes

| Code | Message | Description |
|---|---|---|
| 2004 | Invalid App | App does not exist |
| 3000 | Invalid Form | Form does not exist |
| 3001 | Name Required | Name is required |
| 3005 | Invalid Parameter(s) | Invalid parameter(s) |
| 3041 | Serial No. Limit Exceeded | Too many Serial No. fields |
| 3042 | Invalid Field Alias | Failed to verify field alias |
| 3083 | Widget Limit Exceeded | Too many widgets in form |
| 3091 | Form Name Limit Exceeded | Form name exceeds 100 chars |
| 3092 | Title Limit Exceeded | Title exceeds 100 chars |

## Data Error Codes

| Code | Message | Description |
|---|---|---|
| 4000 | Data Submission Failed | Failed to submit data |
| 4001 | Invalid Data | Data does not exist |
| 4042 | Data Deletion Failed | Failed to delete data |
| 4402 | Aggregation Validation Failed | Failed aggregation calculation |
| 4815 | Invalid Filter | Invalid filter condition |
| 17023 | Batch Update Limit Exceeded | Batch update limit exceeded |
| 17024 | Batch Creation Limit Exceeded | Batch creation limit exceeded |

## Workflow Error Codes

| Code | Message | Description |
|---|---|---|
| 4007 | No Approver | No approver for workflow |
| 4008 | Closed Workflow | Workflow was closed |
| 4009 | No Permission | No permission to approve |
| 4015 | Invalid Approver | Transferred node approver invalid |
| 4016 | Action Failed | Can't transfer to oneself |
| 4025 | No Permission | No permission for workflow |
| 5003 | Invalid Node | Workflow node doesn't exist |
| 5004 | No Approval Comment | No comment for approval |
| 5006 | Transfer Error | Can't transfer to CC node |
| 5008 | Approver Not Found | Can't find the approver |
| 5009 | Invalid Approver | Returned node approver invalid |
| 5011 | Invalid Node | Target node is invalid |
| 5012 | No Signature | No signature for approval |
| 5024 | No Permission | No permission at this node |
| 5025 | Node Completed | Current node has been completed |
| 5026 | Batch Approve Failed | Batch approve failed at task node |
| 5034 | Child Workflow(s) Found | Node has child workflows/plugin nodes |
| 5044 | Child Workflow Error | Child workflow wrongly configured |
| 5045 | Child Workflow Limit | Child workflows exceed 200 |
| 5049 | Transfer Disabled | Transfer not enabled at node |
| 5053 | Action Failed | Invalid initial department for multi-level approval |
| 5056 | Invalid Form | Can't perform workflow ops on non-workflow forms |
| 50004 | Invalid Task | Task doesn't exist |
| 50008 | Action Failed | Workflow processing error |
| 50011 | Invalid Task | Task doesn't exist |
| 50013 | Approver Count Mismatch | Approvers before/after must be same count |
| 50014 | Invalid Approver | No permission to close task |
| 50015 | Approver Not Changed | Approver unchanged |
| 50016 | Invalid instance_id | Instance_id is invalid |
| 50019 | Invalid Action | Action not supported |
| 50021 | Invalid data_id | data_id is invalid |
| 50024 | Empty Approver | Approver can't be empty after adjustment |
| 50025 | Batch Limit Exceeded | Single batch approval limit exceeded |
| 50031 | Invalid data_id | data_id is invalid |
| 50040 | Return Failed | This node can't be returned |
| 50041 | Return Failed | Can't return to current node |
| 50046 | Too Many Pending Tasks | Too many pending sub-workflows |
| 50047 | Workflow Being Migrated | Try again later |
| 50049 | Return Failed | Can't return to Start Node without initiator |
| 50051 | Duplicate Request | Workflow already approved/transferred/returned |
| 50052 | Add Approver Failed | Can't add approver for this node |
| 50053 | No Approver | No approver to add |
| 50054 | No Nesting | Can't add nested approval |
| 50055 | Action Failed | No pre/post-approver added |
| 50056 | Invalid Parent Task | Parent task is lost |
| 50057 | Add Approver Disabled | Feature not enabled |
| 50059 | Approver Added | Member is added as approver |
| 50060 | Invalid Approver | Approver is invalid |
| 50061 | Approver Limit | Can't add more than one approver |
| 50062 | Added Approver Reviewing | Added approver is reviewing |
| 50063 | Batch Submit Failed | Unfinished tasks with added approvers |
| 50064 | Loop Detected | Async sub-workflow forms a loop |
| 50070 | Return Failed | Can't return to plugin node |
| 50071 | Plugin Unavailable | Plugin nodes unavailable in current version |
| 50073 | Return Failed | Can't return to in-progress node |

## Department Error Codes

| Code | Message | Description |
|---|---|---|
| 6000 | Duplicate Sub-Department | Same-name sub-department exists |
| 6001 | Invalid Parent Department | Parent department doesn't exist |
| 6002 | Invalid Department | Department doesn't exist |
| 6003 | Department with Sub-Department(s) | Can't delete with sub-departments |
| 6004 | Department Update Failed | Failed to update |
| 6005 | Department Creation Failed | Failed to create |
| 6006 | Department with Member(s) | Can't delete with members |
| 6010 | Invalid Department ID | Invalid ID format |
| 6011 | Circular Department Relationship | Circular relationship detected |
| 6012 | Invalid Department Name | Name is invalid |
| 6013 | Duplicate Department ID | ID already exists |
| 6014 | No Sub-Department Under Root | Must have at least one sub-dept |
| 6015 | Root Department | Can't delete root department |
| 6017 | Cascading Levels Limit | Cascade levels exceed limit |
| 6019 | Member List Required | Member list is required |
| 6020 | Imported Departments Limit | Single import departments exceeded |
| 6021 | Imported Members Limit | Single import members exceeded |
| 6064 | Duplicate Department | Parent department already exists |

## Plan/Quota Error Codes

| Code | Message | Description |
|---|---|---|
| 7103 | System Limit Exceeded | System suspended, upgrade required |
| 7212 | Data Limit Exceeded | Monthly data limit reached |
| 7216 | Attachment Upload Limit | Contact link creator |
| 7217 | Attachment Upload Limit | Contact business owner |
| 7218 | Attachment Upload Limit | Contact link creator |
| 7219 | Attachment Upload Limit | Contact business owner |
