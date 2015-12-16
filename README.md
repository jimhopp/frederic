# frederic
Web app that tracks clients for a St. Vincent de Paul conference.

[![Build Status](https://travis-ci.org/jimhopp/frederic.svg?branch=master)](https://travis-ci.org/jimhopp/frederic)

This app runs on Google's appengine. See https://cloud.google.com/appengine/docs/go/ for details.

You can run the app locally by downloading the SDK at the link above and cloning this repo. If you want to run the app on
Google's infrastructure you'll need to create your own application at https://console.developers.google.com/ and update app.yaml with
your application name.

The BOOTSTRAP_USER field in app.yaml is used to log into the app initially; there's a user table but it's empty.
