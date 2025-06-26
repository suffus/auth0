Specification of User Actions

A user may perform an action.  Actions are user-instigated events such as sign in, sign out, log in to <resource>, 
install an app, uninstall an app, give a user a permission, revoke a permission, etc.

To represent the actions, we need to add a table to the database to store the possible actions. The table of possible actions is a list of all the actions that a user can perform.  It is a list of strings.  The strings are the names of the actions.  It also has a column containing a string which contains a json array of zero or more permissions (called required_permissions), all of which the user must have in order to be allowed to perform the action.

We also need to modify the AuthenticationLog table to add a column for the id of the action being performed (if any), and another column for JSON representing the effect of the action (named "json_detail"), which will be in a structure that may differ from action to action (but will always be JSON). 


The API supports the authorization and recording of actions by users.  

A POST to the /auth/action/${action_name} API call will do the following:
    (a) look up action_name in the Actions table and retrun a 404 error if it does not exist
    (b) confirm that the device code in the Authorization header of the call authenticates the user and that the user has the required permissions in the permissions field if the action (if any).
    (c) create a row in the AuthenticationLog table which records the action.  For now, the json_detail field will simply be filled with the content of the POST.

An example may be:

POST /auth/action/ssh-login HTTP/1.1
Content-type: application/json
Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj

{resource:"aws-cloud-west/server101", login:"support"}

This will first check that the action 'ssh-login' exists, then identify the user from the Authorization header (maybe returning a 401 at this point).  Then the call will verify that the identified user is permitted to perform the specified action (in this case, "ssh-login") which may (for the sake of argument) require the 'ssh:login' permission for the user (which can be read from the JSON array in the required_permissions column of the Actions table).



