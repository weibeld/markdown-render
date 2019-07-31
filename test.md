# Writing a note-taking app

In this section, you will write an application from scratch.

Later in this course, you will package this app as a Docker image and deploy it to Kubernetes.

The application is a note-taking app — similar to Evernote or Google Keep.

* * *

It allows to format notes with Markdown and include pictures.

Here is how it will look like:

![Adding images and notes in Knote](assets/knote-add-image.gif)

_Now that you know how the app works, it's your turn to code it._

## Bootstrapping the app

You will build the app with JavaScript and Node.js (only server-side JavaScript, no client-side code).

> If at any time you're stuck, you can find the final code of the app [in this repository](https://github.com/learnk8s/knote-js/tree/master/01).

_Let's get started._

Create a directory called `knote` and `cd` into it.

_You will do the front-end stuff first._

Within `knote`, create a folder named `public` and download [Tachyons CSS file](https://raw.githubusercontent.com/learnk8s/knote-js/master/01/public/tachyons.min.css) to `public/tachyons.min.css`:


```terminal|command=1,2|title=bash
mkdir public
wget -P public https://raw.githubusercontent.com/learnk8s/knote-js/master/01/public/tachyons.min.css
```

Next create a folder named `views`:

```terminal|command=1|title=bash
mkdir views
```

And save the following [Pug](https://pugjs.org/api/getting-started.html) template to `views/index.pug`.

```pug|title=index.pug
html
  head
    title= title
    link(rel="stylesheet", href="tachyons.min.css")
  body.ph3.pt0.pb4.mw7.center.sans-serif
    h1.f2.mb0 #[span.gold k]note
    p.f5.mt1.mb4.lh-copy A simple note-taking app.
    form(action="/note" method="POST" enctype="multipart/form-data")
      ol.list.pl0
        li.mv3
          label.f6.b.db.mb2(for="image") Upload an image
          input.f6.link.dim.br1.ba.b--black-20.ph3.pv2.mb2.dib.black.bg-white.pointer(type="file" name="image")
          input.f6.link.dim.br1.ba.bw1.ph3.pv2.mb2.dib.black.bg-white.pointer.ml2(type="submit" value="Upload" name="upload")
        li.mv3
          label.f6.b.db.mb2(for="description") Write your content here
          textarea.f4.db.border-box.hover-black.w-100.measure.ba.b--black-20.pa2.br2.mb2(rows="5" name="description") #{content || ''}
          input.f6.link.dim.br1.ba.bw1.ph3.pv2.mb2.dib.black.bg-white.pointer(type="submit" value="Publish" name="publish")
    if notes.length > 0
      ul.list.pl0
        p.f6.b.db.mb2 Notes
        each note in notes
          li.mv3.bb.bw2.b--light-yellow.bg-washed-yellow.ph4.pv2
            p.measure!= note.description
    else
      p.lh-copy.f6 You don't have any notes yet.
```

This will be the main page of your application.

Pug is a templating engine for HTML that allows generating HTML files in a concise way.

_With the front-end stuff out of the way, let's turn to coding._

Start by initialising a Node.js app:

```terminal|command=1|title=bash
npm init
```

And install the first two dependencies:

```terminal|command=1|title=bash
npm install pug express
```

You've already seen Pug — your app also uses [Express](https://expressjs.com/) to serve the web content.

Now, create an `index.js` file with the following content:

```js|title=index.js
const path = require('path')
const express = require('express')

const app = express()
const port = process.env.PORT || 3000

async function start() {
  app.set('view engine', 'pug')
  app.set('views', path.join(__dirname, 'views'))
  app.use(express.static(path.join(__dirname, 'public')))

  app.listen(port, () => {
    console.log(`App listening on http://localhost:${port}`)
  })
}

start()
```

This is not much more than some standard boilerplate code for an Express app.

It doesn't yet do anything useful.

_But you will change this now by connecting a database._

## Connecting a database

The database will store the notes that users create in your app.

There exist many database solutions, but [MongoDB](https://www.mongodb.com/) is well-suited for your note-taking application because it's easy to set up and doesn't introduce the overhead of a relational database.

So, go on and install the MongoDB package:

```terminal|command=1|title=bash
npm install mongodb
```

Then add the following lines to the beginning of your `index.js` file:

```js|title=index.js
const MongoClient = require('mongodb').MongoClient
const mongoURL = process.env.MONGO_URL || 'mongodb://localhost:27017'
```

Now your code needs to connect to the MongoDB server.

**You have to consider something important here.**

_Your app shouldn't crash when the database isn't ready at the time the app starts up._

Instead, the app should keep retrying to connect to the database until the connection succeeds.

_This is important because later you will delegate running your app to Kubernetes._

Kubernetes expects that application components can be started in any order.

> When you develop an application for Kubernetes, you have to keep Kubernetes considerations in mind from the beginning.

Add the following function to your `index.js` file:

```js|title=index.js
async function initMongo() {
  console.log("Initialising MongoDB...")
  let success = false
  while (!success) {
    try {
      client = await MongoClient.connect(mongoURL, { useNewUrlParser: true });
      success = true
    } catch {
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
  }
  console.log("MongoDB initialised")
  return client.db('dev').collection('notes')
}
```

This function keeps trying to connect to a MongoDB database at the given URL until the connection succeeds.

It then creates a database named `dev` and a collection named `notes`.

> MongoDB collections are like tables in relational databases — lists of items.

Now you have to call the above function inside the `start` function that you already have in your `index.js` file:

```js|title=index.js
async function start() {
  const db = await initMongo() 
  // ...
```

Note how the `await` keyword blocks the execution of the app until the database is ready.

_The next step is to actually use the database._

## Saving and retrieving notes

When the main page of your app loads, two things happen:

- All the existing notes are displayed
- Users can create new notes through an HTML form

_Let's address the displaying of existing notes first._

You need a function to retrieve all notes from the database:

```js|title=index.js
async function retrieveNotes(db) {
  const notes = (await db.find().toArray()).reverse()
  return notes
}
```

The route of the main page is `/`, so you have to add a corresponding route to your Express app that retrieves all the notes from the database and displays them:

Add the following route inside the `start` function that you already have in your `index.js` file:

```js|title=index.js
async function start() {
  // ...
  app.get('/', async (req, res) => {
    res.render('index', { notes: await retrieveNotes(db) })
  })
  // ...
}
```

The above handler calls the `retrieveNotes` function to get all the notes from the database and then passes them into the `index.pug` template.

_Next, let's address the creation of new notes._

First of all, you need a function that saves a single note in the database:

```js|title=index.js
async function saveNote(db, note) {
  await db.insertOne(note)
}
```

The form for creating notes is defined in the `index.pug` template.

Note that it handles both the creation of notes and the uploading of pictures.

You will use [Multer](https://github.com/expressjs/multer), a middle-ware form multi-part form data, to handle the uploaded data.

Go on and install Multer:

```terminal|command=1|title=bash
npm install multer
```

And include it in your code:

```js|title=index.js
const multer = require('multer')
```

The form submits to the `/note` route, so you need to add this route to your Express app:

```js|title=index.js
async function start() {
  // ...
  app.post('/note', multer().none(), async (req, res) => {
    if (!req.body.upload && req.body.description) {
      await saveNote(db, { description: req.body.description })
      res.redirect('/')
    }
  })
  // ...
}
```

The above handler calls the `saveNote` function with the content of the text box, which causes the note to be saved in the database.

It then redirects to the main page, so that the newly created note appears immediately on the screen.

**Your app is functional now (although not yet complete)!**

You can already run your app at this stage.

But to do so, you need to run MongoDB as well.

You can install MongoDB following the instructions in the [official MongoDB documentation](https://docs.mongodb.com/manual/installation/).

Once MongoDB is installed, start a MongoDB server with:

```terminal|command=1|title=bash
mongod
```

Now run your app with:

```terminal|command=1|title=bash
node index.js
```

The app should connect to MongoDB and then listen for requests.

You can access your app on <http://localhost:3000>.

You should see the main page of the app.

Try to create a note — you should see it being displayed on the main page.

_Your app seems to works._

**But it's not yet complete.**

The following things are missing:

- Markdown text is not formatted but just displayed verbatim
- Uploading pictures does not yet work

_Let's fix these issues next._

## Rendering Markdown to HTML

The Markdown notes retrieved from the database should be rendered to HTML so that the web browser can display them properly formatted.

[Marked](https://github.com/markedjs/marked) is an excellent engine for rendering Markdown to HTML.

As always, install the package first:

```terminal|command=1|title=bash
npm install marked
```

And import it in your `index.js` file:

```js|title=index.js
const marked = require('marked')
```

Then, change the `retrieveNotes` function as follows (changed lines are highlighted):

```js|highlight=3-5|title=index.js
async function retrieveNotes(db) {
  const notes = (await db.find().toArray()).reverse()
  return notes.map(it => {
    return { ...it, description: marked(it.description) }
  })
}
```

The new code converts all the notes to HTML before returning them.

Restart the app and access it on <http://localhost:3000>.

**All your notes should now be nicely formatted.**

_This was the first issue, let's tackle the next one._

## Uploading pictures

Your goal is: when a user uploads a picture, the picture should be saved somewhere, and a link to the picture should be inserted in the text box.

This is similar to how adding pictures on StackOverflow works.

> Note that for this to work, the picture upload handler must have access to the text box — this is the reason that picture uploading and note creation are combined in the same form.

For now, the pictures will be stored on the local file system.

You will use Multer to handle the uploaded pictures.

Change the handler for the `/note` route as follows (changed lines are highlighted):

```js|highlight=3,8-14|title=index.js
async function start() {
  // ...
  app.post('/note', multer({ dest: path.join(__dirname, 'public/uploads/') }).single('image'), async (req, res) => {
    if (!req.body.upload && req.body.description) {
      await saveNote(db, { description: req.body.description })
      res.redirect('/')
    }
    else if (req.body.upload && req.file) {
      const link = `/uploads/${encodeURIComponent(req.file.filename)}`
      res.render('index', {
        content: `${req.body.description} ![](${link})`,
        notes: await retrieveNotes(db),
      })
    }
  })
  // ...
}
```

The new code saves uploaded pictures in the `public/uploads` folder in your app directory and inserts a link to the file into the text box.

Restart your app again and access it on <http://localhost:3000>.

Try to upload a picture — you should see a link being inserted in the text box.

And when you publish the note, the picture should be displayed in the rendered note.

**Your app is feature complete now.**

> Note that you can find the complete code for the app in [in this repository](https://github.com/learnk8s/knote-js/tree/master/01).

You just created a simple note-taking app from scratch.

In the next section, you will learn how to package and run your app as a Docker container.
