# CIPTr

Do you work for an MSP and can't stand yet another spreadsheet to enter your devices information into?  
Or are you a solo IT admin and you randomly choose a hostname every time and hope it's not already taken?  

### **C**lient **IP** **Tr**acker is here to help you better manage your infrastructure and focus on what's really important.

> [!NOTE]  
> This project is, at the moment of writing, purely vibecoded. It is used by a real organization to track more than a hunderds network and thousand of devices, but bugs are expected and can be reported.

> [!CAUTION]  
> This project is under heavy development and use in production is not recommended. If you really want to deploy this project make sure your backup are working as intended and are restorable.

## Features
CIPTr gives you a fully featured terminal UI experience to better manage your infrastructure.  
Using hierarchical or absolute view, you can easely navigate across all your devices without loosing track of what matters.

### Views
CIPTr gives you different workflow style, both hierarchical
![Hierarchical](./tapes/browse.gif)
and absolute
![Absolute](./tapes/sections.gif)

### Multiple users with audit
CIPTr natively supports multiple user, each one with its own role, allowing you to limit feature access and record audit trail.  
All of this in a nicely packed login screen

![Authentication](./tapes/authentication.gif)

### Find what you're searching, easely
Every table fully support filtering every field for an even faster search experience/

![Filter](./tapes/filter.gif)

### Custom TUI form for entity creation / updates
No unhandy button or slow graphical web application (yet!).  
Handle all your entity directly inside the terminal UI.

![Form](./tapes/form.gif)


### New version? You're covered
CIPTr can update itself by running a simple command

![version](./tapes/version.gif)


You can start by visiting the [Installation](#installation) section.




## Installation

```sh
curl -sSfL https://raw.githubusercontent.com/guerrieroriccardo/CIPTr/main/scripts/install.sh | sh
```
