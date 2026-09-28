<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
</script>

<DialogPrimitive.Trigger asChild {...$$restProps}>
	<slot />
</DialogPrimitive.Trigger>
