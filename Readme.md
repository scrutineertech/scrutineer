Scrutineer CLI
==============

Scrutineer is a tool to sign and verify commits.
Verification is much more difficult than signing because it involves proper key and trust management.

Scrutineer's backend centralizes key management and allows for easy management of trust relationships.

Install
-------

If you use MacOs, the preferred way is to use 
```zsh
brew tap scrutineertech/scrutineer; brew install scrutineer
```

With brew it is easy to stay up to date.

If you have go installed, you can alternatively use `go install github.com/scrutineertech/scrutineer`.

Getting started
---------------

Login happens via GitHub:

```bash
scrutineer login
```

Follow the steps in the cli. It doesn't even take one minute.

If you like, you can trust the head-author's Scrutineer User "UC2MUTHXY".

```bash
scrutineer trust user UC2MUTHXY
```

Once trusted, you can clone this repository and verify the commits yourself with

```bash
git log --show-signature
```

Trust relationships
-------------------

The heart of Scrutineer's verification logic is a so called Realm.
You define your own Realm.

You can set a start and an end time for a trust relationship.
By default, you trust a given user from now to 365 days from now.

If the user you want to trust has already a history of signing commits, it makes sense to date back your
trust for their signature when they started signing commits that you want to trust.

### How to find User-handles?

You can ask a person for their User-handle. The Handle is also part of every signed commits.
To read the raw commit, you can run `git log --format=raw`.
In most cases it makes sense to have a secure communication channel with the other person to exchange the User-handle.

### Example

The User UC2MUTHXY started working on a repository on 2023-01-01. You want to trust all signed commits from this user.
```bash
scrutineer trust user --start 2023-01-01T00:00:00 UC2MUTHXY
```
This trust relationship will end in 365 days from now. It is good practice to let trust expire and renew it
once it makes sense. Trust should not be granted forever. It would be possible though with the `--end` flag.
