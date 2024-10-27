# GoUpdater

## WIP
This is *extremely* early on in development. If you know go and want to add a feature feel free to PR it.

## What does this do?
This is a wrapper for Minecraft servers that will automatically check for updates from a remote repo (using git tags)
and pull them when a new version is found. A secondary goal will be to sync the mods to the clients, however this does
pose security risks. 

## How to use it?
*You* probably don't have to. But if you are one of the unlucky few who are managing Minecraft Servers by hand, there
are some steps.

### Setting up the configs
The two main configs here are `.env` and `config.json`, which have examples inside the root of the repo. Just set the
repo url metadata in `config.json` and customize your flags. The repo url should point to one that looks like
[JaegyuDev/mc-server-tracked-template](https://github.com/JaegyuDev/mc-server-tracked-template). The repo should basically
just be a regular server minus game data. The flags within config file are generally pretty good though, if not a little
heavy on the RAM side of things, so not much tweaking should be required.  

You will only need the `.env` file for private github repos. This is a planned feature but isn't implemented yet.

### SystemD integration
SystemD integration is one of the biggest focuses of this project. Ideally you should be able to just start your aws
instance and kick back. The wrapper will come in two parts, the SystemD facing part, which will track the most
recent hosted server, as well as start it on reboots and crashes. To set this up just copy `systemd/minecraft-server*`
to your `/etc/systemd/system` directory.

The service facing part of the wrapper will handle
the updates, wrapping the server so you can access it with a berkley socket, and everything else already stated. In this
project's current iteration (v0.1) it's going to be all the same application, however v1.0 will either be two parts, or
handled similar to BusyBox handles it.