# Padsync

Tracking etherpads. Regularly take snapshots of a etherpad and store it in a git repository.

# Background

[The Carpentries](https://carpentries.org/) use etherpads during trainings,
which has some advantages:

* people can *edit collaboratively*, add and correct material during lessons
* people see others are active (m)
* when asked questions, you can see people writing (m)
* it is a snapshot of material covered in a lesson

There are some disadvantages as well:

* minimalistic format when using plaintext (maybe use hackmd, codimd)
* content might get lost
* limited notion of history

# Solution

Regularly export and commit content from etherpad into a git repository.
Multiple pads can be synced into a single repo.

```shell
$ padsync -p https://yourpart.eu/p/example -g git@git.example.com:user/pads.git
```


