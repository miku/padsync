# Padsync

Tracking etherpads. Regularly take snapshots of a etherpad.

# Background

The Carpentries use etherpads during trainings. Advantages:

* people can edit collaboratively, add and correct during lessons
* people see others are active (motivating)
* when asked questions, you can see people writing (motivating)
* it is a snapshot of material covered in a lesson

There are some disadvantages as well:

* minimalistic format when using plaintext (maybe use hackmd, codimd)
* content might get lost
* hard to see development of content

# Solution

Regularly export and commit content from etherpad into a git repository.
Multiple pads can be synced into a single repo.

```shell
$ padsync -p https://yourpart.eu/p/padsync -g git@git.ramenlinux.com:miku/pads.git -n padsync
```


