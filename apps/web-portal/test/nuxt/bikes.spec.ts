import { it, expect, describe, vi, beforeEach } from "vitest";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import Marketplace from "~/pages/marketplace/index.vue";
import BikeDetails from "~/pages/bikes/[id].vue";
import MyBikes from "~/pages/bikes/mine.vue";
import { useBikeRegistration } from "~/composables/useBikeRegistration";

// Create stable mock functions
const mockGet = vi.fn();
const mockPost = vi.fn();

vi.mock("~/composables/useApi", () => ({
  useBikesApi: () => ({
    GET: mockGet,
    POST: mockPost,
  }),
}));

// Mock fetch for binary uploads
global.fetch = vi.fn();

describe("Bikes Marketplace Logic", () => {
  beforeEach(() => {
    mockGet.mockReset();
  });

  it("displays bikes when API returns data", async () => {
    const mockBikes = {
      bikes: [
        {
          id: "1",
          make_model: "Specialized",
          price: 1000,
          status: "for_sale",
          images: [],
        },
      ],
    };
    mockGet.mockResolvedValue({ data: mockBikes, error: null });
    const component = await mountSuspended(Marketplace);
    expect(component.text()).toContain("Specialized");
  });
});

describe("Bike Registration Logic", () => {
  beforeEach(() => {
    mockPost.mockReset();
    vi.mocked(fetch).mockReset();
  });

  it("orchestrates the multi-stage registration flow correctly", async () => {
    const { registerBike } = useBikeRegistration();

    // 1. Mock Metadata Creation
    mockPost.mockResolvedValueOnce({
      data: { id: "new-bike-uuid" },
      error: null,
    });

    // 2. Mock Presigned URL Generation
    mockPost.mockResolvedValueOnce({
      data: { upload_url: "http://s3.test/upload", object_key: "key1" },
      error: null,
    });

    // 3. Mock Binary Upload (fetch)
    vi.mocked(fetch).mockResolvedValueOnce({ ok: true } as any);

    // 4. Mock Confirmation
    mockPost.mockResolvedValueOnce({
      data: { success: true },
      error: null,
    });

    const mockFile = new File(["dummy content"], "bike.jpg", {
      type: "image/jpeg",
    });

    await registerBike(
      {
        make_model: "New Bike",
        year: 2024,
        price: 500,
        serial_number: "SN123",
        location_city: "Berlin",
      },
      [mockFile],
    );

    // Verify Stage 1: Metadata
    expect(mockPost).toHaveBeenCalledWith(
      "/bikes",
      expect.objectContaining({
        body: expect.objectContaining({ make_model: "New Bike" }),
      }),
    );

    // Verify Stage 2: Upload URL
    expect(mockPost).toHaveBeenCalledWith(
      "/bikes/{id}/upload-url",
      expect.objectContaining({
        params: { path: { id: "new-bike-uuid" } },
      }),
    );

    // Verify Stage 3: Binary Fetch
    expect(fetch).toHaveBeenCalledWith(
      "http://s3.test/upload",
      expect.objectContaining({
        method: "PUT",
      }),
    );

    // Verify Stage 4: Confirm
    expect(mockPost).toHaveBeenCalledWith(
      "/bikes/{id}/images/confirm",
      expect.objectContaining({
        body: { object_key: "key1" },
      }),
    );
  });

  it("handles failures during metadata registration", async () => {
    const { registerBike, registrationError } = useBikeRegistration();

    mockPost.mockResolvedValueOnce({
      data: null,
      error: { error: "Duplicate serial number" },
      response: { status: 409 },
    } as any);

    await expect(
      registerBike({ make_model: "Fail", serial_number: "EXISTS" } as any, []),
    ).rejects.toThrow();

    expect(registrationError.value).toContain("Duplicate serial number");
  });
});

describe("Bike Details Logic", () => {
  beforeEach(() => {
    mockGet.mockReset();
  });

  it("displays bike details correctly", async () => {
    const mockBike = {
      id: "123",
      make_model: "Trek Fuel",
      status: "for_sale",
      images: [],
    };
    mockGet.mockResolvedValue({ data: mockBike, error: null });
    const component = await mountSuspended(BikeDetails, {
      route: "/bikes/123",
    });
    expect(component.text()).toContain("Trek Fuel");
  });
});

describe("My Bikes Collection Logic", () => {
  beforeEach(() => {
    mockGet.mockReset();
  });

  it("displays user collection", async () => {
    const mockMyBikes = {
      bikes: [
        { id: "mine-1", make_model: "Giant", status: "registered", images: [] },
      ],
    };
    mockGet.mockResolvedValue({ data: mockMyBikes, error: null });
    const component = await mountSuspended(MyBikes);
    expect(component.text()).toContain("Giant");
  });
});
