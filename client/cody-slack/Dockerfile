# Choose the latest LTS (Long Term Support) version of node as our base image
FROM node:lts

ARG PROJECT_ROOT="."

# Create a directory for our application
WORKDIR /app

# Copy only the build artifact and package.json into our app directory
COPY ${PROJECT_ROOT}/dist ./dist
COPY ${PROJECT_ROOT}/package/package.json ./

# Install only production dependencies
RUN npm install -g pnpm && pnpm install --prod

# The application's default port is picked from the environment
EXPOSE $PORT

# Start the application
CMD [ "node", "dist/app.js" ]
