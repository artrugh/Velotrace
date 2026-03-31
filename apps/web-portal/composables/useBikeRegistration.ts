import type { components } from "@api-contract/.generated/bikes";

export type BikeRegistrationRequest =
  components["schemas"]["handler.RegisterBikeRequest"];
export type Bike = components["schemas"]["models.Bike"];

export const useBikeRegistration = () => {
  const bikesApi = useBikesApi();
  const authToken = useCookie("auth_token");

  const isRegistering = ref(false);
  const registrationError = ref<string | null>(null);
  const registrationProgress = ref(0);

  /**
   * Orchestrates the multi-stage bike registration waterfall:
   * 1. Create Bike Metadata
   * 2. For each image (Parallel):
   *    a. Get Presigned URL
   *    b. Upload Binary to Storage
   *    c. Confirm Upload with API
   */
  const registerBike = async (
    form: BikeRegistrationRequest,
    images: File[],
  ): Promise<Bike> => {
    isRegistering.value = true;
    registrationError.value = null;
    registrationProgress.value = 0;

    try {
      if (!authToken.value) {
        throw new Error("Authentication session expired. Please log in again.");
      }

      const headers = {
        Authorization: `Bearer ${authToken.value}`,
      };

      // --- Stage 1: Metadata Registration ---
      const { data: bike, error: apiError } = await bikesApi.POST("/bikes", {
        body: form,
        headers,
      });

      if (apiError || !bike) {
        const msg = (apiError as any)?.error || "Failed to create bike record.";
        throw new Error(`[Registration] ${msg}`);
      }

      const bikeId = bike.id;
      registrationProgress.value = 10; // First 10% for metadata

      if (images.length === 0) {
        registrationProgress.value = 100;
        return bike;
      }

      // --- Stage 2, 3 & 4: Parallel Image Processing ---
      // We divide the remaining 90% progress by (images * 3 sub-steps)
      const stepWeight = 90 / (images.length * 3);

      const processImage = async (file: File) => {
        // Sub-stage A: Get Presigned URL
        const { data: urlData, error: urlError } = await bikesApi.POST(
          "/bikes/{id}/upload-url",
          {
            params: { path: { id: bikeId } },
            body: { filename: file.name },
            headers,
          },
        );

        if (urlError || !urlData) {
          const msg =
            (urlError as any)?.error ||
            `Could not get upload URL for ${file.name}`;
          throw new Error(`[Auth] ${msg}`);
        }
        registrationProgress.value += stepWeight;

        const { upload_url, object_key } = urlData;

        // Sub-stage B: Binary Upload to Storage
        const uploadResponse = await fetch(upload_url, {
          method: "PUT",
          body: file,
          headers: { "Content-Type": file.type },
        });

        if (!uploadResponse.ok) {
          throw new Error(
            `[Storage] Failed to upload ${file.name}. Storage returned ${uploadResponse.status}.`,
          );
        }
        registrationProgress.value += stepWeight;

        // Sub-stage C: Confirm Upload with API
        const { error: confirmError } = await bikesApi.POST(
          "/bikes/{id}/images/confirm",
          {
            params: { path: { id: bikeId } },
            body: { object_key },
            headers,
          },
        );

        if (confirmError) {
          const msg =
            (confirmError as any)?.error || `Failed to finalize ${file.name}`;
          throw new Error(`[Confirm] ${msg}`);
        }
        registrationProgress.value += stepWeight;

        return object_key;
      };

      // Run all image pipelines in parallel
      await Promise.all(images.map(processImage));

      registrationProgress.value = 100;
      return bike;
    } catch (e: any) {
      const errorMessage = e.message || "An unexpected error occurred.";
      registrationError.value = errorMessage;
      console.error("Bike Registration Flow Failed:", e);
      throw e;
    } finally {
      isRegistering.value = false;
    }
  };

  return {
    registerBike,
    isRegistering,
    registrationError,
    registrationProgress,
  };
};
