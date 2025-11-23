import http from "k6/http";
import {check, sleep} from "k6";

export const options = {
  vus: 20,         
  duration: "30s", 
};

const BASE_URL = "http://localhost:8080";
const TEAM_NAME = "backend-test";

function makePullRequestId(vuId) {
  return (
    "pr-k6-" +
    vuId +
    "-" +
    Date.now() +
    "-" +
    Math.random().toString(16).slice(2)
  );
}

function createTeamIfNeeded() {
  const payload = {
    team_name: TEAM_NAME,
    members: [
      { user_id: "u1", username: "Ivan", is_active: true },
      { user_id: "u2", username: "Vasiliy",   is_active: true },
      { user_id: "u3", username: "Sveta",   is_active: true },
    ],
  };

  const res = http.post(
    `${BASE_URL}/team/add`,
    JSON.stringify(payload),
    {
      headers: { "Content-Type": "application/json" },
    },
  );

  check(res, {
    "team created or already exists": (r) =>
      r.status === 201 || r.status === 409,
  });
}

function createPullRequest(prId) {
  const payload = {
    pull_request_id: prId,
    pull_request_name: "Load Test PR",
    author_id: "u1",
  };

  const res = http.post(
    `${BASE_URL}/pullRequest/create`,
    JSON.stringify(payload),
    {
      headers: { "Content-Type": "application/json" },
    },
  );

  check(res, {
    "create PR: status 201": (r) => r.status === 201,
  });

  return res;
}

function mergePullRequest(prId) {
  const payload = {
    pull_request_id: prId,
  };

  const res = http.post(
    `${BASE_URL}/pullRequest/merge`,
    JSON.stringify(payload),
    {
      headers: { "Content-Type": "application/json" },
    },
  );

  check(res, {
    "merge PR: status 200": (r) => r.status === 200,
  });

  return res;
}

function getReviewForUser(userId) {
  const res = http.get(
    `${BASE_URL}/users/getReview?user_id=${userId}`,
  );

  check(res, {
    "getReview: status 200": (r) => r.status === 200,
  });

  return res;
}

export function setup() {
  createTeamIfNeeded();
  return {};
}

export default function () {
  const prId = makePullRequestId(__VU);

  createPullRequest(prId);

  if (Math.random() < 0.5) {
    mergePullRequest(prId);
  }

  getReviewForUser("u2");

  sleep(0.1);
}