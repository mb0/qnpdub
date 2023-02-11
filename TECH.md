Tech
====

The qnpdub project should provide useful tools and services for amateur musicians to publish music.

The project should only use other open source software and primarily use the [Go programming
language](https://golang.org) to write missing open source tools (because the author and first
contributer are interested in this language).

The following sub projects can be explored and worked on independently, and at a later stage can be
combined to provide additional features.

rman
----

One key service of qnpdub is the discovery and management of song rights. We want to build on
existing datasets of song metadata. This part of the project could be used as local or private
service to manage song licenses for small artists and labels too.

After preliminary search, one can find the [MusicBrainz database project](https://musicbrainz.org)
that manages a public [CC0](http://creativecommons.org/publicdomain/zero/1.0/) licensed dataset of
music meta informations. The MusicBrains project uses the open source [PostgreSQL
database](https://www.postgresql.org/). The dataset provides artists and label credits, dates,
titles, genres and so much more.

This task is ideal for someone who wants to learn writing a regular web service in Go. There already
is a publicly available dataset for postgresql to work with.

The users may want to discover and add songs to their portfolio, a list of songs they indent to dub.
We want to provide a search mask to select songs or releases already in the dataset. We also want
the ability to add new releases that are not yet part of the dataset (e.g. a new jam on youtube).
We want to discover or add the licenses and lists of relevant rights holders for each song. Users
might want to track their progress and correspondence regarding permission or tolerance for their
dubs. Users might want to see and share their portfolio of songs, add entries with links to their
published dubs (derivative works).


For development or later for local deployment we could add a config flag to use a single user.
For a web version we would need a account table for admins, members and users. Some basic session
cookie based email and password authentication can be used in the beginning. This can later be
extended by using a oauth2 library (like [goth](https://github.com/markbates/goth)).

Then we need some database tables to model our own data. Probably some tables to link works to user
accounts, and tables adding details to these entries. This depends on the dataset schema we use and
many small design decisions along the way. I am certain that it makes no sense to specify upfront
how things should be organized into models, as this might be an important part of the exercise.

promo
-----

General public information for visitors explaining the problem and proposed solutions.

This is mostly writing and refining some informational texts.

If we already have Oauth integration we would want to provide simple features to visitors:

 * Leave your email to get notification about major future milestones.
 * Sign a general petition for rights holders to highlight the need.

Both features use only a database table with a reference to a user account for name and email.
And maybe a simple cache of the petition list, that gets invalidated when a visitor adds his name.

We could first use the collaborative features of github to edit markdown files and use a markdown
processor (like [goldmark](https://github.com/yuin/goldmark)) to render the same markdown files on
our server.

plea
----

Personalized and protected landing pages for artists and publishers to explain the project and to
get them on board. The information should be landing page style content. It should be light on text,
with links for more information and a short video introduction to give a personalized overview.
The page should feature at least one example dub as showcase.

Simple interface for:
 * Indicating tolerance, interest or even granting restricted rights to the project.
 * Or artist account registration with pre-filled references to catalogue data to white-list
   specific works of art.
 * Leave feedback and preferred contact information for the project.
 * Leave a quote or testimonial for the project to advertise with.

To actually grant rights on the spot we need to get a media lawyer on board to vet our idea pro
bono, and maybe write up example license agreements.

We may want to feature flag some elements, so we can show small independent artists an interface to
grant rights, but hide it for big labels and commercial focused performers.

I singled out independent artists Jack Conte (from Patreon, Pomplamoose and Scary Pockets) and
Louise Cole (from KNOWER) for these plea attempts, mostly because I already have a collage ready to
showcase what this would look like.

To get prominent support would be an important milestone to get this project underway.
The early scope of this sub project would be custom made mockup, which then gets refined and
integrated with the content management tools and the catalog services.


Other sub-projects
------------------

### oauth sign-on

Most potential users already have a google or soundcloud account. We can use Oauth2 integration to
allow simple account registration with these providers.

### qnpdub widget

The idea is that we mix or synchronize two media streams on the client side in such a way that both
the original work as well as the dub can be published on their own, both get hits and count as view
impression, and the volume can be independently adjusted to some extent (it should not be possible
to mute the dub and only hear the original though, to avoid misuse).

This involves mostly investigating the various web apis for audio and video to build a proof of
concept. Mixing two audio sources by themselves is easy enough. The question is can we mix the audio
of two video sources as well?

The next task would be to integrate the widget with the youtube and soundcloud player apis.


### catalog apis

We my want to integrate with the youtube and soundcloud apis, at least to integrate their players
and to query additional information about links to releases on that platform.

Later we might even cover publications on these platforms in the scope of collaborative dub
projects.

Other potential interesting platforms are mixcloud, bandcamp, apple music and spotify. They probably
have their own api endpoints.

### youtube metadata

If we intend to create video collages, we better match the frame rate for our own video recording.
It seems, that the youtube does not provide this information in their public api.
We can, however, use [yt-dlp](https://github.com/yt-dlp/yt-dlp) to discover youtube video formats
and frame rates. The question remains if this is tolerated by google. We certainly need a worker
pool and some rate limiting for jobs, to not appear rude.

### blender templates

We could investigate the blender file format and think about providing templates for video collages
that are already prepared to match a youtube video duration and framerate, ready to drop in your own
clip and the downloaded youtube video. The templates could come with a default qnpdub branded
background image and some text fades with info about the original song and the dub.

### renderq

A super simple and small bash or go program that manages a blender background rendering queue.
For my discovery CongaDub project I wrote a 33 line version in bash, that already proved helpful.
One could build on the idea and provide some overall progress reporting and job time estimates.

### qnpdub api

We could later provide our own api to query the links and artists of dubs marked as public.
The idea would be that before publishers review copyright violations, they can automatically check
against our system and hopefully drop those claims in the spirit of fair-use.

### project tools

Investigate using git for Ardour and Blender projects with large file support for audio and video
tracks. Both Ardour and Blender are non destructive editors that leave the original recordings
intact. Investigate alternative source control solution used for AV work in the industry.

