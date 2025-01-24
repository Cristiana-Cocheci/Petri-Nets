# Petri Nets for Concurrent Programming

#### Original paper and Rust implementation:

https://arxiv.org/pdf/2208.02900 \
https://github.com/MarshallRawson/nt-petri-net

## Introduction

Petri Nets were created by Carl Petri in 1939 with the initial purpose of modelling chemical reactions. They are now used for `designing concurrent programs`.

A Petri Net is a `weighted bipartite directed graph` with two types of nodes:

1. `Place` Nodes
   - contain any number of tokens
   - have edges from and to transition nodes
   - when an out edge activates, it decreases the token count by its weight
   - when an in edge activates, it increases the token count by its weight
2. `Transition` Nodes
   - do not contain tokens
   - have edges from and to place nodes
   - a transition can only fire if all place nodes that point to it have more tokens that the weight of their respected edges
   - when a transition fires, it decreases the token count of its input places and increases the token count of its output places
   - are not obligated to preserve the total token count in the net

| ![H20 Net Before](/assets/Intro1.png) | ![H2O Net After](/assets/Intro2.png) |
| :-----------------------------------: | :----------------------------------: |
|     _Before the transition fires_     |     _After the transition fires_     |

A state of a Petri net represents the allocation of tokens in places. The state graph is a collection of states and the transitions that cause the respective token distribution

Our Go implementation provides a way to `read` a petri net from a json file, `run` the petri net using the maximum amount of useful threads, and `write` the coresponding state graph to a txt file.

## Work Clusters

Transition places are `partitioned` such that no place node can have out edges to two transitions in different partitions. Each partition is equivalent to a work cluster which is then mapped to a `goroutine`. This partitioning prevents race conditions and deadlocks from happening. The reasoning is that, since `transitions fire sequentially within a work cluster`, no other thread can decrement the input place of a firing transition (incrementing doesn't affect the transition because we only care that the token count is never less than zero).

|    ![Initial petri net](/assets/Clusters1.png)    |
| :-----------------------------------------------: |
| _Petri Net before being split into work clusters_ |

|    ![Split petri net](/assets/Clusters2.png)     |
| :----------------------------------------------: |
| _Petri Net after being split into work clusters_ |

Whenever multiple transitions are enabled (have the required tokens to fire) within a work cluster, we pick one of them at random to actually fire, leading to `nondeterminism`. Each work cluster also has an associated `channel` that notifies it that at least one of its nodes has had its token amount updated. This way, we only check which transitions are enabled if we received an update (as opposed to running this check in an infinite loop). The channels are `closed` whenever a certain total time has passed or the time between transitions has exceeded a given amount.

## Examples

#### Mutex

We can model mutual exclusion with a place node that holds exactly `one token`, the lock's `key`. Once a transition node consumes this token, all others are blocked. When the token is returned, either waiting transitions can fire.

| ![Initial configuration](/assets/Mutex1.png) |
| :------------------------------------------: |
|   _Initial configuration of the Petri Net_   |

| ![Highlighted transitions and tokens](/assets/Mutex2.png) |
| :-------------------------------------------------------: |
|             _Lock1 has enough tokens to fire_             |

|     ![Net after Lock1 fired](/assets/Mutex3.png)      |
| :---------------------------------------------------: |
| _Lock1 fired, Lock2 cannot fire since mutex is empty_ |

|             ![Net after Unlock1 fired](/assets/Mutex4.png)             |
| :--------------------------------------------------------------------: |
| _Unlock1 fired, the mutex has a token and either transitions can fire_ |

#### Dining Philosophers

This problem is automatically solved by the fact that transitions are `triggered atomically`. A philosopher can only start eating once they `take` both forks, which happens once their take transition has enough tokens (one from each fork place node) and fires. Once the philosopher is done eating, they `return` both forks with their respective give transition, so that their neighbours may start eating.

| ![Initial configuration](/assets/Philosophers1.png) |
| :-------------------------------------------------: |
|      _Initial configuration of the Petri Net_       |

| ![Highlighted take transition](/assets/Philosophers2.png) |
| :-------------------------------------------------------: |
|              _Philosopher1 took both forks_               |

Analyzing the work clusters for this petri nets reveals that we can have at most 6 `useful threads`: one for each philosopher's give transition and one global thread managing the take transitions. This result is natural since the code section for allocating resources (forks) need to be concurrent in this problem.

#### Santa Claus

In this concurrency problem, the goal is to chose a group of either 3 threads of a kind (Elfs), either 9 threads of another kind (Reindeers). Once a group has been chosen, some processing occurs and only once the group is released can another group be chosen. This type of problem is easily solved with a petri net's `weighted edges` going into the transitions. We add a place with one `key token` for the mutual exclusion.

| ![Santa Claus solution](/assets/Santa.png) |
| :----------------------------------------: |
|      _Petri Net solving the problem_       |

As opposed to the Dining Philosophers example, this net has a `simple design` for a more `complex problem`.

## Future ideas

#### Colored Petri Nets

In the original paper, the concept of colored petri nets is briefly mentioned. Adding a color for each token and edge encodes `conditional logic` into the petri net (a transition fires only `if` it has enough tokens of the right `color` from a place).

#### Tokens with Information

Currently, each place only holds the number of tokens. However, each token could contain `extra data` (a number, an id, a resource).

#### Transitions as Functions

Right now, a transition takes tokens from certain places and puts tokens in others. We could add extra functionality to these transitions: processing the information in tokens (e.g. taking three tokens each with a number and outputing one token that holds the sum of these numbers), logging messages with `print` or even modelling time interactions in a petri net structure using `sleep`.
