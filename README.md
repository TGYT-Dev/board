## Connecting
You *NEED* IPv6. You can test whether or not you have support for it via [this link](https://test-ipv6.com/)
ssh into the board via the command below.

```
ssh 2a01:4f9:3080:3192::2056 -p 2222
```

## Overview

This is a message board built in a week for the crescent YSWS event run by hackclub.
It is written in GO, my first time doing this so probably a little weird as I am kinda learning as I go along. It is accessed via SSH.
It runs on BubbleTea and Wish. I hope you like it!

## Features
> - SQlite User DB
> - Replies
> - User ID based on SSH Public Key
> - Custom Usernames
> - And More !

## Hosting
I host this on Nest, hack clubs hosting service for teens, and plan on buying a domain from cloudflare.

## Controls
| Key | Action |
|---|---|
| j / k | Move between posts |
| Tab | Switch between posts and replies |
| Ctrl+N | Write a new post or reply |
| Ctrl+S | Send post (press Enter to save the username during setup) |
| Ctrl+E | Cancel writing |
| Ctrl+L | Like or unlike the selected post |
| Ctrl+C | Quit |
