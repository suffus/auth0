Spec for frontend management app

In the frontend app, we now wish to create a frontend management app that supports the following operations:

A login page which will have a device code entry box like the proof of concept which will accept a yubikey (or other device code) and log the user in (or deny access). It will create a new session so that read accesses to the yubiapp will not need a separate device entry per call. On successful login, the app will direct the user to the dashboard page.

A dashboard page that shows:

    A widget which shows the total number of users in the system and the number whose status is "online", alomg with a graph over time of these numbers.
    A widget which gives a history of the actions each user has performed (ordered by time).
    a widget that supports registration of yubikeys and other devices to yourself and others.
    A "message of the day" widget that shows a message that can be updated by a manager.

A device registration page that allows you to:

    Register a yubikey to a user (including creating the Yubikey device for new devices).
    Deregister an existing yubikey.

A users page that allows you to:

    View a filtered set of users and display their status.
    Assign a user to a role.
    Remove a user from a role.

A roles and permissions page that allows you to:

    list the roles and permissions in the system
    add a permission to a role
    remove a permission from a role

An actions page that:

    lists the actions performed on the system filterable by date, user, location, action type, etc.
    allows users to perform specific actions such as signing in and out, going to lunch, showing they are available or unavailable, etc.
    allows managers to perform actions on behalf of other users.

Each page should have a consistent menu bar that allows them to select the organization for which they are looking at, a link to the dashboard page, a link to the users page, a link to the actions page and a link to the help page.
