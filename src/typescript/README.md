# Overview
## index.ts
This is the entrypoint for the app, handle setup after the window loads and determines the auth state of the user

## pages/app.ts
Contains the app logic, initialises/manages state for each section of the app. Any inter-section communication should be added here

## Sections and State
Sections may have state that's synchronised with the server, depending on their purpose. For the sections that do have state, you'll need to write a class that inherits from `BaseState`. 

The section will HAVE A state instance that manages any, requests, caching and/or local storage of server data for the section. This decouples rendering (sections) and state (state instance).

### pages/sections/
Sections attach to the html and manage state for different app functions, the following sections:
- notes
- scheduled
- stack
- tags
- trash

Also have a "state" instance, which manages state for the section.

### state
Manages state for sections

## components
Contains optional components for forms, lists, etc

## ds
Contains any data structures and classes that relate to interacting with those data structure (code here shouldn't do any rendering)

## utility
Contains miscellaneous utility functions